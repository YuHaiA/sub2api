//go:build unit

package service

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/tidwall/gjson"
)

func TestApplyGrokRepeatedToolCallGuardInjectsFailureRecoveryInstruction(t *testing.T) {
	body := buildGrokToolLoopTestBody(t, []grokToolLoopTestAttempt{
		{arguments: `{"cmd":"pwsh -c bad"}`, output: grokToolLoopParserError("first", "0.31")},
		{arguments: `{"cmd":"pwsh -c bad"}`, output: grokToolLoopParserError("second", "0.42")},
		{arguments: `{"cmd":"pwsh -c bad"}`, output: grokToolLoopParserError("third", "0.53")},
	})

	guarded, signal, err := applyGrokRepeatedToolCallGuard(body)
	require.NoError(t, err)
	require.NotNil(t, signal)
	require.Equal(t, "exec_command", signal.ToolName)
	require.Equal(t, 3, signal.RepeatCount)
	require.NotEmpty(t, signal.CallFingerprint)
	require.NotEmpty(t, signal.OutputFingerprint)
	require.True(t, signal.Failed)
	require.Contains(t, gjson.GetBytes(guarded, "instructions").String(), grokToolLoopGuardMarker)
	require.Contains(t, gjson.GetBytes(guarded, "instructions").String(), "Do not call it again with those arguments")
}

func TestApplyGrokRepeatedToolCallGuardInjectsAfterFirstFailure(t *testing.T) {
	body := buildGrokToolLoopTestBody(t, []grokToolLoopTestAttempt{
		{arguments: `{"cmd":"pwsh -c bad"}`, output: grokToolLoopParserError("first", "0.31")},
	})

	guarded, signal, err := applyGrokRepeatedToolCallGuard(body)
	require.NoError(t, err)
	require.NotNil(t, signal)
	require.Equal(t, "exec_command", signal.ToolName)
	require.Equal(t, 1, signal.RepeatCount)
	require.Contains(t, gjson.GetBytes(guarded, "instructions").String(), "explicitly failed")
	require.Contains(t, gjson.GetBytes(guarded, "instructions").String(), "then continue the task")
	require.NotContains(t, gjson.GetBytes(guarded, "instructions").String(), "temporarily unavailable")
}

func TestApplyGrokRepeatedToolCallGuardPreservesExistingInstructions(t *testing.T) {
	body := buildGrokToolLoopTestBody(t, []grokToolLoopTestAttempt{
		{arguments: `{"cmd":"bad"}`, output: grokToolLoopParserError("one", "0.1")},
		{arguments: `{"cmd":"bad"}`, output: grokToolLoopParserError("two", "0.2")},
		{arguments: `{"cmd":"bad"}`, output: grokToolLoopParserError("three", "0.3")},
	})
	var request map[string]any
	require.NoError(t, json.Unmarshal(body, &request))
	request["instructions"] = "existing system instruction"
	body, err := json.Marshal(request)
	require.NoError(t, err)

	guarded, signal, err := applyGrokRepeatedToolCallGuard(body)
	require.NoError(t, err)
	require.NotNil(t, signal)
	require.Contains(t, gjson.GetBytes(guarded, "instructions").String(), "existing system instruction")
	require.Contains(t, gjson.GetBytes(guarded, "instructions").String(), grokToolLoopGuardMarker)
}

func TestApplyGrokRepeatedToolCallGuardResetsChangedFailures(t *testing.T) {
	tests := []struct {
		name     string
		attempts []grokToolLoopTestAttempt
	}{
		{
			name: "arguments change",
			attempts: []grokToolLoopTestAttempt{
				{arguments: `{"cmd":"bad-1"}`, output: grokToolLoopParserError("one", "0.1")},
				{arguments: `{"cmd":"bad-1"}`, output: grokToolLoopParserError("two", "0.2")},
				{arguments: `{"cmd":"bad-2"}`, output: grokToolLoopParserError("three", "0.3")},
			},
		},
		{
			name: "failure changes",
			attempts: []grokToolLoopTestAttempt{
				{arguments: `{"cmd":"bad"}`, output: "Process exited with code 1\nOutput:\nfirst error"},
				{arguments: `{"cmd":"bad"}`, output: "Process exited with code 1\nOutput:\nsecond error"},
				{arguments: `{"cmd":"bad"}`, output: "Process exited with code 1\nOutput:\nthird error"},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, signal, err := applyGrokRepeatedToolCallGuard(buildGrokToolLoopTestBody(t, test.attempts))
			require.NoError(t, err)
			require.NotNil(t, signal)
			require.Equal(t, 1, signal.RepeatCount)
		})
	}
}

func TestApplyGrokRepeatedToolCallGuardInjectsForUnchangedSuccessfulOutput(t *testing.T) {
	body := buildGrokToolLoopTestBody(t, []grokToolLoopTestAttempt{
		{toolName: "read_file", arguments: `{"path":"a.go"}`, output: grokToolLoopSuccess("first", "0.1", "same content")},
		{toolName: "read_file", arguments: `{"path":"a.go"}`, output: grokToolLoopSuccess("second", "0.2", "same content")},
		{toolName: "read_file", arguments: `{"path":"a.go"}`, output: grokToolLoopSuccess("third", "0.3", "same content")},
	})

	guarded, signal, err := applyGrokRepeatedToolCallGuard(body)
	require.NoError(t, err)
	require.NotNil(t, signal)
	require.False(t, signal.Failed)
	require.Equal(t, 3, signal.RepeatCount)
	require.Contains(t, gjson.GetBytes(guarded, "instructions").String(), "materially unchanged successful output")
	require.Contains(t, gjson.GetBytes(guarded, "instructions").String(), "continue the task")
}

func TestApplyGrokRepeatedToolCallGuardIgnoresChangedSuccessfulOutput(t *testing.T) {
	body := buildGrokToolLoopTestBody(t, []grokToolLoopTestAttempt{
		{toolName: "read_file", arguments: `{"path":"a.go"}`, output: "first content"},
		{toolName: "read_file", arguments: `{"path":"a.go"}`, output: "second content"},
	})

	guarded, signal, err := applyGrokRepeatedToolCallGuard(body)
	require.NoError(t, err)
	require.Nil(t, signal)
	require.JSONEq(t, string(body), string(guarded))
}

func TestSuppressGrokRepeatedToolRemovesOnlyTarget(t *testing.T) {
	body := []byte(`{"tools":[{"type":"function","name":"exec_command"},{"type":"function","name":"apply_patch"}],"tool_choice":{"type":"function","name":"exec_command"},"input":[]}`)

	suppressed, changed, err := suppressGrokRepeatedTool(body, "exec_command")
	require.NoError(t, err)
	require.True(t, changed)
	require.Len(t, gjson.GetBytes(suppressed, "tools").Array(), 1)
	require.Equal(t, "apply_patch", gjson.GetBytes(suppressed, "tools.0.name").String())
	require.Equal(t, "auto", gjson.GetBytes(suppressed, "tool_choice").String())
}

func TestApplyGrokRepeatedToolCallGuardMarksSuppressionThreshold(t *testing.T) {
	attempts := make([]grokToolLoopTestAttempt, grokRepeatedToolSuppressionThreshold)
	for index := range attempts {
		attempts[index] = grokToolLoopTestAttempt{
			arguments: `{"cmd":"pwsh -c bad"}`,
			output:    grokToolLoopParserError(fmt.Sprintf("chunk-%d", index), fmt.Sprintf("0.%d", index)),
		}
	}

	guarded, signal, err := applyGrokRepeatedToolCallGuard(buildGrokToolLoopTestBody(t, attempts))
	require.NoError(t, err)
	require.NotNil(t, signal)
	require.Equal(t, grokRepeatedToolSuppressionThreshold, signal.RepeatCount)
	require.Contains(t, gjson.GetBytes(guarded, "instructions").String(), "temporarily unavailable")
}

type grokToolLoopTestAttempt struct {
	toolName  string
	arguments string
	output    string
}

func buildGrokToolLoopTestBody(t *testing.T, attempts []grokToolLoopTestAttempt) []byte {
	t.Helper()
	input := make([]any, 0, len(attempts)*2)
	for index, attempt := range attempts {
		toolName := attempt.toolName
		if toolName == "" {
			toolName = "exec_command"
		}
		callID := fmt.Sprintf("call-%d", index)
		input = append(input,
			map[string]any{"type": "function_call", "call_id": callID, "name": toolName, "arguments": attempt.arguments},
			map[string]any{"type": "function_call_output", "call_id": callID, "output": attempt.output},
		)
	}
	body, err := json.Marshal(map[string]any{
		"model": "grok-4.6",
		"tools": []any{map[string]any{"type": "function", "name": "exec_command"}},
		"input": input,
	})
	require.NoError(t, err)
	return body
}

func grokToolLoopParserError(chunkID, wallTime string) string {
	return fmt.Sprintf("Chunk ID: %s\nWall time: %s seconds\nProcess exited with code 1\nOriginal token count: 60\nOutput:\nParserError: Unexpected token '120' in expression or statement.", chunkID, wallTime)
}

func grokToolLoopSuccess(chunkID, wallTime, output string) string {
	return fmt.Sprintf("Chunk ID: %s\nWall time: %s seconds\nProcess exited with code 0\nOriginal token count: 20\nOutput:\n%s", chunkID, wallTime, output)
}
