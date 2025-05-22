import { useState, useEffect, useCallback } from 'react';
import { Toaster } from "sonner"
import { Button } from "@/components/ui/button"
import { Settings, Music2, Terminal, DownloadIcon, ListIcon, HistoryIcon } from "lucide-react"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import DownloadManager from "@/components/DownloadManager"
import QueueManager from "@/components/QueueManager"
import HistoryManager from "@/components/HistoryManager"
import ConfigurationDialog from "@/components/ConfigurationDialog"
import QuickHelp from "@/components/QuickHelp"
import { SKIP_TO_CONTENT_CLASS, setupKeyboardUserDetection, applyKeyboardUserStyles } from '@/lib/accessibility';
import { type DownloadJob, type ProgressUpdate, type SSEEvent } from "@/types/api"

const AppHeader = () => {
  const [configOpen, setConfigOpen] = useState(false)
  
  return (
    <header className="flex items-center justify-between mb-8">
      <div className="flex items-center gap-3">
        <div className="rounded-md bg-purple-500/20 p-2">
          <Music2 className="h-5 w-5 text-purple-400" />
        </div>
        <h1 className="text-2xl font-bold bg-gradient-to-r from-purple-400 to-pink-400 text-transparent bg-clip-text">
          nugs-dl
        </h1>
      </div>
      <Button 
        variant="ghost" 
        size="icon" 
        className="h-10 w-10 text-gray-400 hover:text-white hover:bg-gray-800/50"
        onClick={() => setConfigOpen(true)}
        aria-label="Open settings"
        aria-haspopup="dialog"
      >
        <Settings className="h-5 w-5" />
      </Button>
      <ConfigurationDialog open={configOpen} onOpenChange={setConfigOpen} />
    </header>
  );
};

const AppFooter = () => {
  return (
    <footer className="mt-8 pt-6 border-t border-gray-800">
      <div className="flex items-center justify-between text-sm text-gray-400">
        <div className="flex items-center gap-2">
          <Terminal className="h-4 w-4" aria-hidden="true" />
          <span>Nugs-Downloader Web UI</span>
        </div>
        <span>v1.0.0</span>
      </div>
    </footer>
  );
};

const MainContentTabs = () => {
    const [activeTab, setActiveTab] = useState("download")
    const [jobs, setJobs] = useState<Record<string, DownloadJob>>({})
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)

    const switchToQueue = () => {
        setActiveTab("queue")
    }



    // Handle progress updates from SSE
    const handleProgressUpdate = useCallback((update: ProgressUpdate) => {
        console.log("[App] handleProgressUpdate called with:", update)
        
        setJobs(prevJobs => {
            const jobId = update.jobId
            const existingJob = prevJobs[jobId]
            if (!existingJob) {
                console.warn(`Received progress for unknown job ID: ${jobId}`)
                return prevJobs
            }
            
            const updatedJob = { ...existingJob }
            if (update.status) updatedJob.status = update.status
            if (update.currentFile) updatedJob.currentFile = update.currentFile
            
            // Preserve the job title - don't let progress updates override it
            // The title should be set by the backend when processing starts
            
            // Store track information
            if (update.currentTrack) updatedJob.currentTrack = update.currentTrack
            if (update.totalTracks) updatedJob.totalTracks = update.totalTracks
            
            // Use track-based progress if available, otherwise fall back to percentage
            if (update.currentTrack && update.totalTracks) {
                // Calculate progress based on completed tracks
                updatedJob.progress = Math.round((update.currentTrack - 1) / update.totalTracks * 100)
                console.log(`[App] Track-based progress: ${update.currentTrack}/${update.totalTracks} = ${updatedJob.progress}%`)
            } else {
                updatedJob.progress = update.percentage < 0 ? -1 : Math.round(update.percentage)
            }
            
            updatedJob.speedBps = update.speedBps
            if (update.status === 'failed' && update.message) {
                updatedJob.errorMessage = update.message
            } else if (update.status !== 'failed') {
                updatedJob.errorMessage = undefined
            }
            const nowStr = new Date().toISOString()
            if (update.status === 'processing' && !updatedJob.startedAt) updatedJob.startedAt = nowStr
            if ((update.status === 'complete' || update.status === 'failed') && !updatedJob.completedAt) {
                updatedJob.completedAt = nowStr
                if (update.status === 'complete') updatedJob.progress = 100
            }
            
            return { ...prevJobs, [jobId]: updatedJob }
        })
    }, [])



    // Load initial jobs and set up SSE - moved to App level
    useEffect(() => {
        let isMounted = true
        let eventSource: EventSource | null = null

        const fetchInitialJobs = async () => {
            console.log("[App] Fetching initial jobs...")
            setIsLoading(true)
            setError(null)
            try {
                const response = await fetch('/api/downloads')
                if (!response.ok) {
                    throw new Error(`Failed to fetch initial jobs: ${response.statusText}`)
                }
                const initialJobs: DownloadJob[] = await response.json()
                if (isMounted) {
                    const initialJobMap: Record<string, DownloadJob> = {}
                    initialJobs.forEach(job => {
                         initialJobMap[job.id] = job 
                    })
                    setJobs(initialJobMap)
                    setError(null)
                }
            } catch (err: unknown) { 
                if (isMounted) {
                     console.error("Error fetching initial jobs:", err)
                     const errorMessage = err instanceof Error ? err.message : 'Failed to load job queue initially.'
                     setError(errorMessage)
                 }
            } finally { 
                if (isMounted) {
                    setIsLoading(false)
                }
            }
        }

        const connectSSE = () => {
            if (!isMounted) return
            console.log("[App] Connecting to SSE stream...")
            eventSource = new EventSource('/api/status-stream')

            // Listen for specific event types
            eventSource.addEventListener('jobAdded', (event) => {
                if (!isMounted) return
                try {
                    const sseEvent: SSEEvent = JSON.parse(event.data)
                    console.log("[App] Received jobAdded event:", sseEvent.data)
                    const newJob = sseEvent.data as DownloadJob
                    setJobs(prevJobs => ({ ...prevJobs, [newJob.id]: newJob }))
                } catch (e) {
                    console.error("Failed to parse jobAdded event:", e, "Data:", event.data)
                }
            })

            eventSource.addEventListener('progressUpdate', (event) => {
                if (!isMounted) return
                try {
                    const sseEvent: SSEEvent = JSON.parse(event.data)
                    console.log("[App] Received progressUpdate event:", sseEvent.data)
                    handleProgressUpdate(sseEvent.data as ProgressUpdate)
                } catch (e) {
                    console.error("Failed to parse progressUpdate event:", e, "Data:", event.data)
                }
            })

            eventSource.addEventListener('jobStatusUpdate', (event) => {
                if (!isMounted) return
                try {
                    const sseEvent: SSEEvent = JSON.parse(event.data)
                    console.log("[App] Received jobStatusUpdate event:", sseEvent.data)
                    const updatedJob = sseEvent.data as DownloadJob
                    setJobs(prevJobs => ({ ...prevJobs, [updatedJob.id]: updatedJob }))
                } catch (e) {
                    console.error("Failed to parse jobStatusUpdate event:", e, "Data:", event.data)
                }
            })

            eventSource.addEventListener('open', () => {
                console.log("[App] SSE connection opened")
            })

            eventSource.onerror = (err) => {
                console.error("SSE Error:", err)
                if (isMounted && eventSource?.readyState === EventSource.CLOSED) {
                    console.log("[App] SSE connection closed, attempting reconnect in 2 seconds...")
                    setTimeout(() => {
                        if (isMounted) {
                            connectSSE()
                        }
                    }, 2000)
                }
            }

            // Handle page visibility changes to reconnect when tab becomes active
            const handleVisibilityChange = () => {
                if (!document.hidden && isMounted && eventSource?.readyState === EventSource.CLOSED) {
                    console.log("[App] Tab became active, reconnecting SSE...")
                    connectSSE()
                }
            }

            document.addEventListener('visibilitychange', handleVisibilityChange)

            return () => {
                document.removeEventListener('visibilitychange', handleVisibilityChange)
            }
        }

        fetchInitialJobs().then(() => {
            const cleanup = connectSSE()
            return cleanup
        })

        return () => {
            console.log("[App] Cleaning up SSE connection.")
            isMounted = false
            eventSource?.close()
        }
    }, [handleProgressUpdate])

    return (
        <div className="w-full max-w-4xl mx-auto space-y-8">
            <section className="text-center space-y-6">
                <h1 className="text-4xl md:text-6xl font-bold leading-tight tracking-tight">
                    Download from <span className="bg-gradient-to-r from-purple-400 via-pink-400 to-purple-600 text-transparent bg-clip-text animate-gradient-x">nugs.net</span> with ease
                </h1>
                <p className="text-xl text-gray-300 max-w-3xl mx-auto leading-relaxed">
                    A powerful tool to download albums, videos, and livestreams from nugs.net in your preferred quality.
                </p>
            </section>

            <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
                <div className="flex justify-center mb-6">
                    <TabsList className="bg-gray-800/50 backdrop-blur-sm border border-gray-700">
                        <TabsTrigger 
                            value="download" 
                            className="flex items-center gap-2 px-6 py-2 text-sm data-[state=active]:bg-gray-700 data-[state=active]:text-purple-300"
                        >
                            <DownloadIcon className="h-4 w-4 text-purple-400 data-[state=active]:text-purple-300" />
                            Download
                        </TabsTrigger>
                        <TabsTrigger 
                            value="queue" 
                            className="flex items-center gap-2 px-6 py-2 text-sm data-[state=active]:bg-gray-700 data-[state=active]:text-purple-300"
                        >
                            <ListIcon className="h-4 w-4 text-purple-400 data-[state=active]:text-purple-300" />
                            Queue
                        </TabsTrigger>
                        <TabsTrigger 
                            value="history" 
                            className="flex items-center gap-2 px-6 py-2 text-sm data-[state=active]:bg-gray-700 data-[state=active]:text-purple-300"
                        >
                            <HistoryIcon className="h-4 w-4 text-purple-400 data-[state=active]:text-purple-300" />
                            History
                        </TabsTrigger>
                    </TabsList>
                </div>
                
                <TabsContent value="download" className="space-y-6">
                    <div className="bg-gray-800/60 backdrop-blur-sm border border-gray-700/50 rounded-xl p-8 shadow-xl shadow-purple-500/10 hover:shadow-purple-500/20 transition-all duration-300">
                        <DownloadManager onStartDownload={switchToQueue} />
                    </div>
                </TabsContent>
                
                <TabsContent value="queue" className="space-y-6">
                    <div className="bg-gray-800/60 backdrop-blur-sm border border-gray-700/50 rounded-xl p-8 shadow-xl shadow-purple-500/10 hover:shadow-purple-500/20 transition-all duration-300">
                        <QueueManager jobs={jobs} isLoading={isLoading} error={error} />
                    </div>
                </TabsContent>
                
                <TabsContent value="history" className="space-y-6">
                    <div className="bg-gray-800/60 backdrop-blur-sm border border-gray-700/50 rounded-xl p-8 shadow-xl shadow-purple-500/10 hover:shadow-purple-500/20 transition-all duration-300">
                        <HistoryManager />
                    </div>
                </TabsContent>
            </Tabs>
            
            <QuickHelp />
        </div>
    );
}

function App() {
  // Set up keyboard navigation detection
  useEffect(() => {
    setupKeyboardUserDetection();
    applyKeyboardUserStyles();
    
    // Add an announcement when the app loads for screen readers
    const announceAppReady = () => {
      const announcer = document.createElement('div');
      announcer.id = 'app-announcer';
      announcer.className = 'sr-only';
      announcer.setAttribute('aria-live', 'polite');
      announcer.textContent = 'Nugs Downloader application loaded. Use tab to navigate.';
      document.body.appendChild(announcer);
      
      // Remove the announcement after it's been read
      setTimeout(() => {
        if (announcer.parentNode) {
          announcer.parentNode.removeChild(announcer);
        }
      }, 3000);
    };
    
    announceAppReady();
  }, []);

  return (
    <div className="min-h-screen bg-gradient-to-b from-gray-900 to-black text-white">
      {/* Skip to content link for keyboard users */}
      <a href="#main-content" className={SKIP_TO_CONTENT_CLASS}>
        Skip to main content
      </a>
      
      <main 
        id="main-content" 
        className="container mx-auto px-4 py-8"
        tabIndex={-1} // Makes the main content focusable by skip link
      >
        <AppHeader />
        <MainContentTabs />
        <AppFooter />
      </main>
      
      <Toaster richColors closeButton theme="dark" />
    </div>
  )
}

export default App
