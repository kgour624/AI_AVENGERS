export interface Expert { id: string; name: string; description?: string }
export interface ToolInfo { name: string; description?: string; inputSchema: any }
export interface AskExpertResult { answer: string; citations?: string[]; expertName?: string }
