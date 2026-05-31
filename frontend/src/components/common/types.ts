/**
 * Common component types
 */

export interface Column {
  key: string
  label: string
  sortable?: boolean
  class?: string
  width?: string
  minWidth?: string
  maxWidth?: string
  formatter?: (value: any, row: any) => string
}
