package service

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	grokRepeatedFailedToolWarningThreshold  = 1
	grokRepeatedSuccessToolWarningThreshold = 2
	grokRepeatedToolSuppressionThreshold    = 5
	grokToolLoopGuardMarker                 = "[Gateway tool-loop guard]"
)

var (
	grokToolNonZeroProcessExitPattern = regexp.MustCompile(`(?im)\b(?:process|command|script)\s+exited with (?:code|status)\s+([1-9][0-9]*)\b`)
	grokToolNonZeroExitCodePattern    = regexp.MustCompile(`(?im)\bexit code\s*[:=]?\s*([1-9][0-9]*)\b`)
)

type grokRepeatedToolCall struct {
	ToolName          string
	RepeatCount       int
	CallFingerprint   string
	OutputFingerprint string
	Failed            bool
}

type grokToolCallHistoryItem struct {
	name        string
	fingerprint string
}

func applyGrokRepeatedToolCallGuard(body []byte) ([]byte, *grokRepeatedToolCall, error) {
	if !json.Valid(body) {
		return body, nil, nil
	}

	requestBody, err := decodeGrokToolLoopRequest(body)
	if err != nil {
		return body, nil, err
	}
	signal := detectGrokRepeatedToolCall(requestBody)
	if signal == nil {
		return body, nil, nil
	}

	var guardInstruction string
	if signal.Failed && signal.RepeatCount == 1 {
		guardInstruction = fmt.Sprintf(
			"%s The latest call to tool %q explicitly failed. Do not repeat it with identical arguments. Inspect the error and change the arguments, quoting, tool, or implementation, then continue the task. If no safe alternative exists, explain the blocker instead of retrying.",
			grokToolLoopGuardMarker,
			signal.ToolName,
		)
	} else if signal.Failed {
		guardInstruction = fmt.Sprintf(
			"%s The tool %q has returned the same explicit failure %d consecutive times with identical arguments. Do not call it again with those arguments. Inspect the latest error and change the arguments, quoting, tool, or implementation, then continue the task. If no safe alternative exists, explain the blocker instead of retrying.",
			grokToolLoopGuardMarker, signal.ToolName, signal.RepeatCount,
		)
	} else {
		guardInstruction = fmt.Sprintf(
			"%s The tool %q has returned materially unchanged successful output %d consecutive times for identical arguments. The latest result is already available in the conversation history. Do not call it again with those arguments; use that result to make progress on the remaining task. If the result is insufficient, change the arguments or use another tool, then continue the task.",
			grokToolLoopGuardMarker, signal.ToolName, signal.RepeatCount,
		)
	}
	if signal.RepeatCount >= grokRepeatedToolSuppressionThreshold {
		guardInstruction += " This tool is temporarily unavailable for this response; continue with another available tool or complete the remaining work without it."
	}

	if existing, ok := requestBody["instructions"].(string); ok && strings.TrimSpace(existing) != "" {
		if !strings.Contains(existing, grokToolLoopGuardMarker) {
			requestBody["instructions"] = existing + "\n\n" + guardInstruction
		}
	} else {
		requestBody["instructions"] = guardInstruction
	}

	rebuilt, err := marshalOpenAIUpstreamJSON(requestBody)
	if err != nil {
		return body, nil, fmt.Errorf("encode Grok tool-loop guard: %w", err)
	}
	return rebuilt, signal, nil
}

func detectGrokRepeatedToolCall(requestBody map[string]any) *grokRepeatedToolCall {
	input, ok := requestBody["input"].([]any)
	if !ok || len(input) == 0 {
		return nil
	}

	pending := make(map[string]grokToolCallHistoryItem)
	var latestPending grokToolCallHistoryItem
	var previous grokRepeatedToolCall
	for _, rawItem := range input {
		item, ok := rawItem.(map[string]any)
		if !ok {
			continue
		}
		itemType := strings.TrimSpace(grokToolLoopString(item["type"]))
		if isGrokToolLoopCallType(itemType) {
			call, ok := grokToolLoopCall(item)
			if !ok {
				continue
			}
			latestPending = call
			if callID := strings.TrimSpace(grokToolLoopString(item["call_id"])); callID != "" {
				pending[callID] = call
			}
			continue
		}
		if !isCodexToolCallOutputItemType(itemType) {
			continue
		}

		call := latestPending
		if callID := strings.TrimSpace(grokToolLoopString(item["call_id"])); callID != "" {
			if matched, exists := pending[callID]; exists {
				call = matched
				delete(pending, callID)
			}
		}
		if call.name == "" || call.fingerprint == "" {
			continue
		}

		outputFingerprint, failed, comparable := grokToolOutputFingerprint(item["output"])
		if !comparable {
			previous = grokRepeatedToolCall{}
			continue
		}
		if previous.CallFingerprint == call.fingerprint &&
			previous.OutputFingerprint == outputFingerprint && previous.Failed == failed {
			previous.RepeatCount++
		} else {
			previous = grokRepeatedToolCall{
				ToolName:          call.name,
				RepeatCount:       1,
				CallFingerprint:   call.fingerprint,
				OutputFingerprint: outputFingerprint,
				Failed:            failed,
			}
		}
	}
	warningThreshold := grokRepeatedSuccessToolWarningThreshold
	if previous.Failed {
		warningThreshold = grokRepeatedFailedToolWarningThreshold
	}
	if previous.RepeatCount < warningThreshold {
		return nil
	}
	return &previous
}

func isGrokToolLoopCallType(itemType string) bool {
	switch itemType {
	case "function_call", "custom_tool_call", "local_shell_call", "mcp_tool_call":
		return true
	default:
		return false
	}
}

func grokToolLoopCall(item map[string]any) (grokToolCallHistoryItem, bool) {
	name := strings.TrimSpace(grokToolLoopString(item["name"]))
	if name == "" {
		return grokToolCallHistoryItem{}, false
	}
	if namespace := strings.TrimSpace(grokToolLoopString(item["namespace"])); namespace != "" {
		name = namespace + "." + name
	}

	argumentValue, exists := firstGrokToolLoopValue(item, "arguments", "input", "action", "command", "cmd")
	if !exists {
		argumentValue = grokToolLoopCallFallback(item)
	}
	canonicalArguments := canonicalGrokToolLoopValue(argumentValue)
	return grokToolCallHistoryItem{
		name:        name,
		fingerprint: hashGrokToolLoopValue(name + "\x00" + canonicalArguments),
	}, true
}

func firstGrokToolLoopValue(item map[string]any, keys ...string) (any, bool) {
	for _, key := range keys {
		if value, ok := item[key]; ok {
			return value, true
		}
	}
	return nil, false
}

func grokToolLoopCallFallback(item map[string]any) map[string]any {
	fallback := make(map[string]any)
	for key, value := range item {
		switch key {
		case "type", "id", "call_id", "name", "namespace", "status", "internal_chat_message_metadata_passthrough":
			continue
		default:
			fallback[key] = value
		}
	}
	return fallback
}

func canonicalGrokToolLoopValue(value any) string {
	if text, ok := value.(string); ok {
		trimmed := strings.TrimSpace(text)
		if json.Valid([]byte(trimmed)) {
			var decoded any
			decoder := json.NewDecoder(strings.NewReader(trimmed))
			decoder.UseNumber()
			if decoder.Decode(&decoded) == nil {
				if encoded, err := json.Marshal(decoded); err == nil {
					return string(encoded)
				}
			}
		}
		return trimmed
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return fmt.Sprintf("%T", value)
	}
	return string(encoded)
}

func grokToolFailureFingerprint(output any) (string, bool) {
	failed := grokToolOutputExplicitlyFailed(output)
	text := grokToolOutputText(output)
	if !failed {
		failed = grokToolNonZeroProcessExitPattern.MatchString(text) ||
			grokToolNonZeroExitCodePattern.MatchString(text) ||
			strings.HasPrefix(strings.TrimSpace(text), "Script failed")
	}
	if !failed {
		return "", false
	}

	canonical := canonicalGrokToolFailureText(text)
	if canonical == "" {
		canonical = "explicit-tool-failure"
	}
	return hashGrokToolLoopValue(canonical), true
}

func grokToolOutputFingerprint(output any) (string, bool, bool) {
	if fingerprint, failed := grokToolFailureFingerprint(output); failed {
		return fingerprint, true, true
	}
	if grokToolOutputPending(output) {
		return "", false, false
	}

	canonical := canonicalGrokToolOutputText(grokToolOutputText(output))
	if canonical == "" {
		canonical = "empty-tool-output"
	}
	return hashGrokToolLoopValue(canonical), false, true
}

func grokToolOutputPending(output any) bool {
	object, ok := output.(map[string]any)
	if ok {
		switch strings.ToLower(strings.TrimSpace(grokToolLoopString(object["status"]))) {
		case "pending", "running", "in_progress":
			return true
		}
	}
	text := strings.ToLower(strings.TrimSpace(grokToolOutputText(output)))
	return strings.HasPrefix(text, "script running with session id") ||
		strings.HasPrefix(text, "process running with session id")
}

func grokToolOutputExplicitlyFailed(output any) bool {
	object, ok := output.(map[string]any)
	if !ok {
		return false
	}
	if isError, ok := object["isError"].(bool); ok && isError {
		return true
	}
	if isError, ok := object["is_error"].(bool); ok && isError {
		return true
	}
	if status := strings.ToLower(strings.TrimSpace(grokToolLoopString(object["status"]))); status == "error" || status == "failed" {
		return true
	}
	for _, key := range []string{"exit_code", "exitCode", "status_code"} {
		if value, exists := object[key]; exists && grokToolLoopNonZeroNumber(value) {
			return true
		}
	}
	return false
}

func grokToolLoopNonZeroNumber(value any) bool {
	switch number := value.(type) {
	case json.Number:
		parsed, err := strconv.ParseInt(number.String(), 10, 64)
		return err == nil && parsed != 0
	case float64:
		return number != 0
	case float32:
		return number != 0
	case int:
		return number != 0
	case int64:
		return number != 0
	case string:
		parsed, err := strconv.ParseInt(strings.TrimSpace(number), 10, 64)
		return err == nil && parsed != 0
	default:
		return false
	}
}

func grokToolOutputText(output any) string {
	if text, ok := output.(string); ok {
		return text
	}
	encoded, err := json.Marshal(output)
	if err != nil {
		return fmt.Sprintf("%v", output)
	}
	return string(encoded)
}

func canonicalGrokToolFailureText(text string) string {
	return canonicalGrokToolOutputText(text)
}

func canonicalGrokToolOutputText(text string) string {
	lines := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	kept := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		lower := strings.ToLower(trimmed)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(lower, "chunk id:") ||
			strings.HasPrefix(lower, "wall time:") ||
			strings.HasPrefix(lower, "process exited with code ") ||
			strings.HasPrefix(lower, "command exited with code ") ||
			strings.HasPrefix(lower, "script exited with code ") ||
			strings.HasPrefix(lower, "process exited with status ") ||
			strings.HasPrefix(lower, "command exited with status ") ||
			strings.HasPrefix(lower, "script exited with status ") ||
			strings.HasPrefix(lower, "original token count:") ||
			lower == "output:" {
			continue
		}
		kept = append(kept, trimmed)
	}
	return strings.Join(kept, "\n")
}

func hashGrokToolLoopValue(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:8])
}

func decodeGrokToolLoopRequest(body []byte) (map[string]any, error) {
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var requestBody map[string]any
	if err := decoder.Decode(&requestBody); err != nil {
		return nil, fmt.Errorf("decode Grok tool-loop request: %w", err)
	}
	return requestBody, nil
}

func suppressGrokRepeatedTool(body []byte, toolName string) ([]byte, bool, error) {
	requestBody, err := decodeGrokToolLoopRequest(body)
	if err != nil {
		return body, false, err
	}
	tools, ok := requestBody["tools"].([]any)
	if !ok || len(tools) == 0 {
		return body, false, nil
	}

	filtered := make([]any, 0, len(tools))
	removed := false
	for _, rawTool := range tools {
		tool, ok := rawTool.(map[string]any)
		if !ok || grokToolDefinitionName(tool) != toolName {
			filtered = append(filtered, rawTool)
			continue
		}
		removed = true
	}
	if !removed {
		return body, false, nil
	}
	requestBody["tools"] = filtered
	if choice, ok := requestBody["tool_choice"].(map[string]any); ok && grokToolDefinitionName(choice) == toolName {
		if len(filtered) == 0 {
			requestBody["tool_choice"] = "none"
		} else {
			requestBody["tool_choice"] = "auto"
		}
	}

	rebuilt, err := marshalOpenAIUpstreamJSON(requestBody)
	if err != nil {
		return body, false, fmt.Errorf("encode suppressed Grok tool request: %w", err)
	}
	return rebuilt, true, nil
}

func grokToolDefinitionName(tool map[string]any) string {
	name := strings.TrimSpace(grokToolLoopString(tool["name"]))
	if namespace := strings.TrimSpace(grokToolLoopString(tool["namespace"])); namespace != "" && name != "" {
		return namespace + "." + name
	}
	return name
}

func grokToolLoopString(value any) string {
	text, _ := value.(string)
	return text
}
