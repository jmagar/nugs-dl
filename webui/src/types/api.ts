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
  title?: string;         
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
  // Track information
  currentTrack?: number;
  totalTracks?: number;

  // Fields apparently returned by /api/downloads/history but missing in type def
  type?: 'album' | 'video' | 'livestream' | 'playlist'; // From HistoryItemProps
  url?: string; // The primary URL of the downloaded item, distinct from originalUrl if processing changes it
  path?: string; // Filesystem path where the item was saved
  sizeFormatted?: string; // Human-readable size (e.g., "1.2 GB")
  format?: string; // e.g., "FLAC", "MP4"
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
  // Track-based progress information
  currentTrack?: number;  // Current track number (1-based)
  totalTracks?: number;   // Total number of tracks
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