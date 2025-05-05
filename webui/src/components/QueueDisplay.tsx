import { useState, useEffect, useCallback } from 'react';
import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table";
import { Progress } from "@/components/ui/progress";
import { Badge } from "@/components/ui/badge";
import { Skeleton } from "@/components/ui/skeleton";
import { Tooltip, TooltipContent, TooltipProvider, TooltipTrigger } from "@/components/ui/tooltip";
import { type DownloadJob, type JobStatus, type ProgressUpdate, type SSEEvent } from "@/types/api";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { toast } from "sonner";
import { Button } from "@/components/ui/button";
import { AlertDialog, AlertDialogAction, AlertDialogCancel, AlertDialogContent, AlertDialogDescription, AlertDialogFooter, AlertDialogHeader, AlertDialogTitle, AlertDialogTrigger } from "@/components/ui/alert-dialog";
import { Trash2, ClipboardCopy } from "lucide-react";

// Simple byte formatter
function formatBytes(bytes: number, decimals = 2): string {
  if (bytes === 0) return '0 Bytes';
  const k = 1024;
  const dm = decimals < 0 ? 0 : decimals;
  const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return parseFloat((bytes / Math.pow(k, i)).toFixed(dm)) + ' ' + sizes[i];
}

export function QueueDisplay() {
  const [jobs, setJobs] = useState<Record<string, DownloadJob>>({});
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [removingJobId, setRemovingJobId] = useState<string | null>(null);

  const handleProgressUpdate = useCallback((update: ProgressUpdate) => {
    setJobs(prevJobs => {
      const jobId = update.jobId;
      const existingJob = prevJobs[jobId];
      if (!existingJob) {
        console.warn(`Received progress for unknown job ID: ${jobId}`);
        return prevJobs;
      }
      const updatedJob = { ...existingJob };
      if (update.status) updatedJob.status = update.status;
      if (update.message) console.log(`[Progress Job ${jobId}] ${update.message}`);
      if (update.currentFile) updatedJob.currentFile = update.currentFile;
      updatedJob.progress = update.percentage;
      updatedJob.speedBps = update.speedBps;
      if (update.status === 'failed' && update.message) {
        updatedJob.errorMessage = update.message;
      } else if (update.status !== 'failed') {
        updatedJob.errorMessage = undefined;
      }
      const nowStr = new Date().toISOString();
      if (update.status === 'processing' && !updatedJob.startedAt) updatedJob.startedAt = nowStr;
      if ((update.status === 'complete' || update.status === 'failed') && !updatedJob.completedAt) {
        updatedJob.completedAt = nowStr;
        if (update.status === 'complete') updatedJob.progress = 100;
      }
      return { ...prevJobs, [jobId]: updatedJob };
    });
  }, []);

  const handleJobAdded = useCallback((job: DownloadJob) => {
    setJobs(prevJobs => {
      return { ...prevJobs, [job.id]: job };
    });
  }, []);

  useEffect(() => {
    let isMounted = true;
    let eventSource: EventSource | null = null;

    const fetchInitialJobs = async () => {
      console.log("[QueueDisplay] Fetching initial jobs...");
      setIsLoading(true);
      setError(null);
      try {
        const response = await fetch('/api/downloads');
        if (!response.ok) {
          throw new Error(`Failed to fetch initial jobs: ${response.statusText}`);
        }
        const initialJobs: DownloadJob[] = await response.json();
        if (isMounted) {
          const initialJobMap: Record<string, DownloadJob> = {};
          initialJobs.forEach(job => {
             initialJobMap[job.id] = job; 
          });
          setJobs(initialJobMap);
          setError(null);
        }
      } catch (err: unknown) { 
        if (isMounted) {
             console.error("Error fetching initial jobs:", err);
             const errorMessage = err instanceof Error ? err.message : 'Failed to load job queue initially.';
             setError(errorMessage);
             toast.error(errorMessage);
         }
      } finally { 
        if (isMounted) {
            setIsLoading(false);
        }
      }
    };

    const connectSSE = () => {
        if (!isMounted) return;
        console.log("[QueueDisplay] Connecting to SSE stream...");
        eventSource = new EventSource('/api/status-stream');

        eventSource.onmessage = null;

        eventSource.addEventListener('message', (event) => {
            if (!isMounted) return;
            try {
                const sseEvent: SSEEvent = JSON.parse(event.data);
                
                switch (sseEvent.type) {
                    case 'jobAdded':
                        handleJobAdded(sseEvent.data as DownloadJob);
                        break;
                    case 'progressUpdate':
                        handleProgressUpdate(sseEvent.data as ProgressUpdate);
                        break;
                    default:
                        console.warn("Received unknown SSE event type:", sseEvent.type);
                }
            } catch (e) {
                console.error("Failed to parse SSE event data:", e, "Data:", event.data);
            }
        });

        eventSource.onerror = (err) => {
            console.error("SSE Error:", err);
            if (isMounted) {
                setError('Real-time connection error. May reconnect automatically...');
                toast.error('Lost connection to status updates.');
                eventSource?.close(); 
            }
        };
    }

    fetchInitialJobs().then(connectSSE);

    return () => {
      console.log("[QueueDisplay] Cleaning up SSE connection.");
      isMounted = false;
      eventSource?.close(); 
    };
  }, [handleJobAdded, handleProgressUpdate]);

  const getStatusBadgeVariant = (status: JobStatus): "default" | "secondary" | "destructive" | "outline" => {
    switch (status) {
      case "processing": return "default";
      case "queued": return "secondary";
      case "complete": return "outline";
      case "failed": return "destructive";
      default: return "secondary";
    }
  };

  const jobList = Object.values(jobs).sort((a, b) => new Date(a.createdAt).getTime() - new Date(b.createdAt).getTime());

  const handleRemoveJob = async (jobId: string) => {
    setRemovingJobId(jobId);
    try {
      const response = await fetch(`/api/downloads/${jobId}`, {
        method: 'DELETE',
      });

      if (!response.ok) {
        let errorMsg = `Failed to remove job (status: ${response.status})`;
        try {
          const errorData = await response.json();
          errorMsg = errorData.error || errorMsg;
        } catch {
          errorMsg = response.statusText || errorMsg;
        } 
        throw new Error(errorMsg);
      }

      setJobs(prevJobs => {
        const newJobs = { ...prevJobs };
        delete newJobs[jobId];
        return newJobs;
      });
      toast.success(`Job ${jobId.substring(0,8)}... removed successfully.`);

    } catch (err: unknown) {
      console.error("Error removing job:", err);
      const errorMessage = err instanceof Error ? err.message : 'Could not remove job';
      toast.error(`Failed to remove job: ${errorMessage}`);
    } finally {
      setRemovingJobId(null);
    }
  };

  // --- Share URL Handler --- 
  const handleShareUrl = async (url: string) => {
    try {
        if (!navigator.clipboard) {
            throw new Error('Clipboard API not available.');
        }
        await navigator.clipboard.writeText(url);
        toast.success("Original URL copied to clipboard!");
    } catch (err) {
        console.error("Failed to copy URL:", err);
        const errorMsg = err instanceof Error ? err.message : 'Could not copy URL';
        toast.error(`Failed to copy: ${errorMsg}`);
    }
  };

  return (
    <Card className="w-full mt-6">
      <CardHeader>
        <CardTitle>Download Queue</CardTitle>
        {error && <p className="text-sm font-medium text-destructive pt-2">Error loading queue: {error}</p>} 
      </CardHeader>
      <CardContent>
        <TooltipProvider delayDuration={300}>
          <Table>
            <TableCaption>{jobList.length === 0 && !isLoading ? "The queue is empty." : isLoading ? "Loading queue..." : "Current download jobs."}</TableCaption>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[60px]">Art</TableHead>
                <TableHead className="w-[100px] hidden sm:table-cell">Job ID</TableHead>
                <TableHead>First URL</TableHead>
                <TableHead>Status</TableHead>
                <TableHead className="w-[150px]">Progress</TableHead>
                <TableHead className="w-[100px] hidden md:table-cell">Speed</TableHead>
                <TableHead className="hidden lg:table-cell">Current File</TableHead>
                <TableHead className="hidden sm:table-cell">Error</TableHead>
                <TableHead className="w-[80px] text-right">Actions</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? (
                Array.from({ length: 3 }).map((_, index) => (
                  <TableRow key={`skeleton-${index}`}>
                    <TableCell><Skeleton className="h-10 w-10 rounded-sm" /></TableCell>
                    <TableCell><Skeleton className="h-4 w-[80px]" /></TableCell>
                    <TableCell><Skeleton className="h-4 w-[200px]" /></TableCell>
                    <TableCell><Skeleton className="h-6 w-[70px] rounded-full" /></TableCell> 
                    <TableCell><Skeleton className="h-4 w-[100px]" /></TableCell>
                    <TableCell><Skeleton className="h-4 w-[80px]" /></TableCell>
                    <TableCell><Skeleton className="h-4 w-[150px]" /></TableCell>
                    <TableCell><Skeleton className="h-4 w-[100px]" /></TableCell>
                    <TableCell className="text-right"><Skeleton className="h-8 w-8" /></TableCell>
                  </TableRow>
                ))
              ) : jobList.length === 0 ? (
                <TableRow>
                  <TableCell colSpan={9} className="h-24 text-center text-muted-foreground">
                    No download jobs in the queue.
                  </TableCell>
                </TableRow>
              ) : (
                jobList.map((job) => {
                  const isRemovable = job.status === 'queued' || job.status === 'failed' || job.status === 'complete';
                  const isCurrentlyRemoving = removingJobId === job.id;
                  return (
                    <TableRow key={job.id}>
                      <TableCell>
                        {job.artworkUrl ? (
                          <img
                            src={job.artworkUrl}
                            alt="Artwork"
                            className="h-10 w-10 object-cover rounded-sm"
                            width={40}
                            height={40}
                            loading="lazy"
                          />
                        ) : (
                          <div className="h-10 w-10 bg-muted rounded-sm flex items-center justify-center text-muted-foreground text-xs">N/A</div>
                        )}
                      </TableCell>
                      <TableCell className="font-mono text-xs hidden sm:table-cell">{job.id.substring(0, 8)}...</TableCell>
                      <TableCell className="font-medium truncate max-w-[150px] sm:max-w-xs">{job.originalUrl}</TableCell>
                      <TableCell>
                        <Badge variant={getStatusBadgeVariant(job.status)}>{job.status}</Badge>
                      </TableCell>
                      <TableCell>
                        {(job.status === 'processing' || job.status === 'complete') && job.progress >= 0 ? (
                          <div className="relative w-full">
                             <Progress value={job.progress < 0 ? 0 : job.progress} className="w-full h-4" />
                             <span className="absolute inset-0 flex items-center justify-center text-[10px] font-medium text-primary-foreground mix-blend-difference"> 
                                {job.progress < 0 ? 'N/A' : `${Math.round(job.progress)}%`}
                             </span>
                          </div>
                        ) : (
                          '-'
                        )}
                      </TableCell>
                      <TableCell className="text-xs hidden md:table-cell">
                        {job.status === 'processing' && job.speedBps > 0 ? `${formatBytes(job.speedBps)}/s` : '-'}
                      </TableCell>
                      <TableCell className="text-xs truncate max-w-[150px] hidden lg:table-cell">
                        {job.status === 'processing' ? job.currentFile : '-'}
                      </TableCell>
                      <TableCell className="text-xs text-destructive truncate max-w-[100px] hidden sm:table-cell">
                        {job.errorMessage && (
                          <Tooltip> 
                             <TooltipTrigger asChild> 
                                <span className="cursor-default">{job.errorMessage}</span> 
                             </TooltipTrigger> 
                             <TooltipContent> 
                                <p>{job.errorMessage}</p> 
                             </TooltipContent> 
                          </Tooltip> 
                        )}
                      </TableCell>
                      <TableCell className="text-right">
                        {isRemovable ? (
                          <AlertDialog>
                            <AlertDialogTrigger asChild>
                              <Button
                                variant="ghost"
                                size="icon"
                                disabled={isCurrentlyRemoving}
                                aria-label="Remove job"
                              >
                                {isCurrentlyRemoving ? (
                                  <span className="animate-spin">⏳</span>
                                ) : (
                                  <Trash2 className="h-4 w-4" />
                                )}
                              </Button>
                            </AlertDialogTrigger>
                            <AlertDialogContent>
                              <AlertDialogHeader>
                                <AlertDialogTitle>Are you sure?</AlertDialogTitle>
                                <AlertDialogDescription>
                                  This action cannot be undone. This will permanently remove the job "{job.id.substring(0,8)}..." ({job.originalUrl}) from the queue.
                                </AlertDialogDescription>
                              </AlertDialogHeader>
                              <AlertDialogFooter>
                                <AlertDialogCancel>Cancel</AlertDialogCancel>
                                <AlertDialogAction onClick={() => handleRemoveJob(job.id)}>
                                  Remove
                                </AlertDialogAction>
                              </AlertDialogFooter>
                            </AlertDialogContent>
                          </AlertDialog>
                        ) : (
                          <span className="text-xs text-muted-foreground">-</span>
                        )}
                        {job.status === 'complete' && (
                           <Button 
                             variant="ghost" 
                             size="icon" 
                             onClick={() => handleShareUrl(job.originalUrl)}
                             aria-label="Copy original URL" 
                             className="ml-1"
                           > 
                             <ClipboardCopy className="h-4 w-4" />
                           </Button>
                        )}
                        {!isRemovable && job.status !== 'complete' && (
                             <span className="text-xs text-muted-foreground">-</span> 
                        )}
                      </TableCell>
                    </TableRow>
                  );
                })
              )}
            </TableBody>
          </Table>
        </TooltipProvider>
      </CardContent>
    </Card>
  );
} 