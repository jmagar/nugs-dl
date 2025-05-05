// Manually converted types from pkg/api/types.go
// TODO: Consider using type generation tools later

export type JobStatus =
  | 'queued'
  | 'processing'
  | 'complete'
  | 'failed';

export interface AppConfig {
  email: string;
  password: string;
  format: number;
  videoFormat: number;
  outPath: string;
  token: string;
  useFfmpegEnvVar: boolean;
}

export interface DownloadOptions {
  forceVideo: boolean;
  skipVideos: boolean;
  skipChapters: boolean;
}

export interface DownloadJob {
  id: string;             
  originalUrl: string;    
  options: DownloadOptions; 
  status: JobStatus;       
  errorMessage?: string;   
  createdAt: string;       
  startedAt?: string;      
  completedAt?: string;    
  progress: number;       
  currentFile?: string;   
  speedBps: number;       
  artworkUrl?: string;    
}

export interface AddDownloadRequest {
  urls: string[];
  options: DownloadOptions;
}

export interface AddDownloadResponseItem {
    url:   string;
    jobId?: string;
    error?: string;
}

export interface ProgressUpdate {
  jobId: string;        
  status?: JobStatus;     
  message?: string;       
  currentFile?: string;   
  percentage: number;    
  speedBps: number;       
  bytesDownloaded: number;
  totalBytes: number;
}

// --- SSE Event Structure ---

// Matches Go type
export type SSEEventType =
  | 'jobAdded'
  | 'progressUpdate'
  | 'message'; // Added generic message type

// Matches Go type
export interface SSEEvent {
  type: SSEEventType;
  data: unknown; // Use unknown and type assertion/checking in listener
} 