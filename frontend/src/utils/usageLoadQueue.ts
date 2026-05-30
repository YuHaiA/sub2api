/**
 * Usage request scheduler.
 *
 * Account usage cells can mount in large batches when the admin account page
 * is opened or restored. Keep these requests bounded so route changes do not
 * flood the browser main thread or backend connection pool.
 */

import type { Account } from '@/types'

const MAX_CONCURRENT_USAGE_REQUESTS = 4

type QueueTask<T> = {
  run: () => Promise<T>
  resolve: (value: T | PromiseLike<T>) => void
  reject: (reason?: unknown) => void
}

const queue: QueueTask<unknown>[] = []
let activeCount = 0

function drainQueue() {
  while (activeCount < MAX_CONCURRENT_USAGE_REQUESTS && queue.length > 0) {
    const task = queue.shift()
    if (!task) return

    activeCount += 1
    task.run()
      .then(task.resolve)
      .catch(task.reject)
      .finally(() => {
        activeCount -= 1
        drainQueue()
      })
  }
}

/**
 * Schedule a usage fetch with a small global concurrency limit.
 */
export function enqueueUsageRequest<T>(
  _account: Account,
  fn: () => Promise<T>
): Promise<T> {
  return new Promise<T>((resolve, reject) => {
    queue.push({
      run: fn,
      resolve: resolve as QueueTask<unknown>['resolve'],
      reject
    })
    drainQueue()
  })
}
