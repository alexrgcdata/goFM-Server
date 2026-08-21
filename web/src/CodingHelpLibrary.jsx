// Preserved for later migration into the independent OpenBridge project.
// This file is intentionally not imported by the goFM admin panel.
export const openBridgeCodingExamples = {
  dataapi: { find: `await fetch('/api/filemaker/find', { method: 'POST' })`, get: `await fetch('/api/filemaker/record/RECORD_ID')`, create: `await fetch('/api/filemaker/record', { method: 'POST' })` },
  odata: { find: `await fetch('/api/odata/customers')`, get: `await fetch('/api/odata/customers/RECORD_ID')`, create: `await fetch('/api/odata/customers', { method: 'POST' })` }
}
