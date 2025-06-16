import { useState } from 'react'
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import QueueItem, { QueueItemProps } from "./QueueItem"
import { toast } from 'sonner'
import { type DownloadJob } from "@/types/api"
import { 
  Search, 
  Pause, 
  Play, 
  X,
  ListIcon,
  Loader2
} from "lucide-react"

interface QueueManagerProps {
  onAddDownload: (url: string) => Promise<void>; // Added to pass down to QueueItem
  jobs: Record<string, DownloadJob>
  isLoading: boolean
  error: string | null
}

const QueueManager = ({ jobs, isLoading, error, onAddDownload }: QueueManagerProps) => {
  const [searchQuery, setSearchQuery] = useState('')
  const [activeTab, setActiveTab] = useState('all')

  // Helper function to convert DownloadJob to QueueItemProps
  const convertJobToQueueItem = (job: DownloadJob): QueueItemProps => {
    return {
      id: job.id,
      title: job.title || extractTitleFromUrl(job.originalUrl, job.currentFile),
      type: 'album', // Default to album, could be enhanced based on URL analysis
      status: job.status as QueueItemProps['status'],
      progress: job.progress < 0 ? -1 : job.progress, // Handle unknown progress, keep decimal precision
      format: 'FLAC', // Default format
      size: 'Unknown', // Would need to be added to DownloadJob type
      speed: job.speedBps > 0 ? formatSpeed(job.speedBps) : undefined,
      eta: undefined, // Would need to be calculated
      errorMessage: job.errorMessage,
      currentTrack: job.currentTrack,
      totalTracks: job.totalTracks,
      artworkUrl: job.artworkUrl,
      originalUrl: job.originalUrl, // Added to pass to QueueItem
      onAddDownload: onAddDownload // Added to pass to QueueItem
    }
  }

  // Extract title from URL (enhanced implementation)
  const extractTitleFromUrl = (url: string, currentFile?: string): string => {
    // If we have currentFile info, try to extract artist/album from it
    if (currentFile) {
      // currentFile might be like "01. Track Name.flac" 
      // We could get better info from the folder structure if available
      const cleanFile = currentFile.replace(/^\d+\.\s*/, '').replace(/\.[^.]*$/, '')
      if (cleanFile && cleanFile !== 'Unknown' && cleanFile.length > 3) {
        return cleanFile
      }
    }

    try {
      const urlObj = new URL(url)
      const pathParts = urlObj.pathname.split('/').filter(part => part.length > 0)
      
      // Handle different nugs.net URL patterns
      if (urlObj.hostname.includes('nugs.net')) {
        if (pathParts.includes('release') && pathParts.length >= 2) {
          const releaseId = pathParts[pathParts.indexOf('release') + 1]
          return `Release ${releaseId}` // More descriptive than just the ID
        }
        if (pathParts.includes('artist') && pathParts.length >= 2) {
          const artistId = pathParts[pathParts.indexOf('artist') + 1]
          return `Artist ${artistId}`
        }
        if (pathParts.includes('playlist') && pathParts.length >= 2) {
          const playlistId = pathParts[pathParts.indexOf('playlist') + 1]
          return `Playlist ${playlistId}`
        }
      }
      
      // Fallback: use the last meaningful part of the path
      const lastPart = pathParts[pathParts.length - 1]
      if (lastPart && lastPart.length > 0) {
        // If it's just a number, make it more descriptive
        if (/^\d+$/.test(lastPart)) {
          return `Download ${lastPart}`
        }
        // Otherwise clean it up
        return lastPart.replace(/-/g, ' ').replace(/\b\w/g, l => l.toUpperCase())
      }
      
      return 'Unknown Download'
    } catch {
      return 'Invalid URL'
    }
  }

  // Format speed for display
  const formatSpeed = (speedBps: number): string => {
    if (speedBps < 1024) return `${speedBps} B/s`
    if (speedBps < 1024 * 1024) return `${(speedBps / 1024).toFixed(1)} KB/s`
    return `${(speedBps / (1024 * 1024)).toFixed(1)} MB/s`
  }

  // Calculate stats from jobs
  const getStats = () => {
    const jobList = Object.values(jobs)
    return {
      all: jobList.length,
      downloading: jobList.filter(job => job.status === 'processing').length,
      queued: jobList.filter(job => job.status === 'queued').length,
      paused: 0, // Not implemented in backend yet
      error: jobList.filter(job => job.status === 'failed').length,
      completed: jobList.filter(job => job.status === 'complete').length
    }
  }

  const stats = getStats()



  // Handle job removal - placeholder for now
  const handleRemoveJob = async (jobId: string) => {
    // TODO: Add job removal logic back with proper state management
    console.log("Remove job:", jobId)
    toast.info("Job removal temporarily disabled during refactor")
  }

  const handlePauseAll = () => {
    toast.info('Pause all functionality not yet implemented')
  }

  const handleResumeAll = () => {
    toast.info('Resume all functionality not yet implemented')
  }

  const handleCancelAll = () => {
    toast.info('Cancel all functionality not yet implemented')
  }

  // Filter items based on active tab and search query
  const filteredJobs = Object.values(jobs).filter(job => {
    // Map job status to tab names
    const statusMap: Record<string, string> = {
      'processing': 'downloading',
      'queued': 'queued', 
      'complete': 'completed',
      'failed': 'error'
    }
    
    const mappedStatus = statusMap[job.status] || job.status
    
    // First filter by tab
    if (activeTab !== 'all' && mappedStatus !== activeTab) {
      return false
    }
    
    // Then filter by search query if present
    if (searchQuery.trim() !== '') {
      const title = extractTitleFromUrl(job.originalUrl)
      return title.toLowerCase().includes(searchQuery.toLowerCase()) ||
             job.originalUrl.toLowerCase().includes(searchQuery.toLowerCase())
    }
    
    return true
  })

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <ListIcon className="h-5 w-5 text-purple-400" />
          <h2 className="text-lg font-medium">Queue Manager</h2>
        </div>
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-purple-400" />
          <span className="ml-2 text-gray-300">Loading queue...</span>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <ListIcon className="h-5 w-5 text-purple-400" />
          <h2 className="text-lg font-medium">Queue Manager</h2>
        </div>
        <div className="text-center py-12 space-y-4">
          <div className="mx-auto w-16 h-16 bg-red-500/20 rounded-full flex items-center justify-center">
            <X className="h-8 w-8 text-red-400" />
          </div>
          <div className="space-y-2">
            <h3 className="text-lg font-medium text-gray-300">Failed to load queue</h3>
            <p className="text-sm text-gray-400 max-w-md mx-auto">{error}</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <ListIcon className="h-5 w-5 text-purple-400" />
        <h2 className="text-lg font-medium">Queue Manager</h2>
      </div>

      {/* Search Bar */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
        <Input
          placeholder="Search downloads..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="pl-10 bg-gray-900/50 border-gray-600 focus:border-purple-400 focus:ring-1 focus:ring-purple-400"
        />
      </div>

      {/* Action Buttons */}
      <div className="flex gap-2">
        <Select value={activeTab} onValueChange={setActiveTab}>
          <SelectTrigger className="w-[160px] bg-gray-900/50 border-gray-600">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Downloads</SelectItem>
            <SelectItem value="downloading">Downloading</SelectItem>
            <SelectItem value="queued">Queued</SelectItem>
            <SelectItem value="completed">Completed</SelectItem>
            <SelectItem value="error">Error</SelectItem>
          </SelectContent>
        </Select>
        
        <Button 
          variant="outline" 
          size="sm" 
          className="flex items-center gap-1.5 border-gray-600 hover:border-gray-500"
          onClick={handlePauseAll}
        >
          <Pause className="h-4 w-4" />
          Pause
        </Button>
        <Button 
          variant="outline" 
          size="sm" 
          className="flex items-center gap-1.5 border-gray-600 hover:border-gray-500"
          onClick={handleResumeAll}
        >
          <Play className="h-4 w-4" />
          Resume
        </Button>
        <Button 
          variant="outline" 
          size="sm" 
          className="flex items-center gap-1.5 border-gray-600 hover:border-gray-500 text-red-400 hover:text-red-300"
          onClick={handleCancelAll}
        >
          <X className="h-4 w-4" />
          Cancel
        </Button>
      </div>

      {stats.all > 0 && (
        <>
          {/* Summary Row */}
          <div className="flex items-center gap-6 py-2 px-3 rounded-lg bg-gray-800/30 border border-gray-700 text-sm">
            <div className="flex items-center gap-2">
              <span className="text-gray-300">Total:</span>
              <span className="font-medium">{stats.all}</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 bg-blue-400 rounded-full"></div>
              <span className="text-blue-400">{stats.downloading} Downloading</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 bg-gray-400 rounded-full"></div>
              <span className="text-gray-400">{stats.queued} Queued</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 bg-green-400 rounded-full"></div>
              <span className="text-green-400">{stats.completed} Completed</span>
            </div>
            <div className="flex items-center gap-2">
              <div className="w-2 h-2 bg-red-400 rounded-full"></div>
              <span className="text-red-400">{stats.error} Error</span>
            </div>
          </div>

          {/* Tabs */}
          <div className="flex gap-1 p-1 rounded-lg bg-gray-800/30 border border-gray-700">
            <button
              onClick={() => setActiveTab('all')}
              className={`px-4 py-2 rounded-md text-sm transition-colors ${
                activeTab === 'all' 
                  ? 'bg-gray-700 text-white' 
                  : 'text-gray-300 hover:text-white hover:bg-gray-800/50'
              }`}
            >
              All ({stats.all})
            </button>
            <button
              onClick={() => setActiveTab('downloading')}
              className={`px-4 py-2 rounded-md text-sm transition-colors ${
                activeTab === 'downloading' 
                  ? 'bg-gray-700 text-white' 
                  : 'text-gray-300 hover:text-white hover:bg-gray-800/50'
              }`}
            >
              Active ({stats.downloading})
            </button>
            <button
              onClick={() => setActiveTab('queued')}
              className={`px-4 py-2 rounded-md text-sm transition-colors ${
                activeTab === 'queued' 
                  ? 'bg-gray-700 text-white' 
                  : 'text-gray-300 hover:text-white hover:bg-gray-800/50'
              }`}
            >
              Queued ({stats.queued})
            </button>
            <button
              onClick={() => setActiveTab('completed')}
              className={`px-4 py-2 rounded-md text-sm transition-colors ${
                activeTab === 'completed' 
                  ? 'bg-gray-700 text-white' 
                  : 'text-gray-300 hover:text-white hover:bg-gray-800/50'
              }`}
            >
              Completed ({stats.completed})
            </button>
            <button
              onClick={() => setActiveTab('error')}
              className={`px-4 py-2 rounded-md text-sm transition-colors ${
                activeTab === 'error' 
                  ? 'bg-gray-700 text-white' 
                  : 'text-gray-300 hover:text-white hover:bg-gray-800/50'
              }`}
            >
              Error ({stats.error})
            </button>
          </div>
        </>
      )}

      {/* Queue Items */}
      <div className="space-y-3">
        {filteredJobs.length > 0 ? (
          filteredJobs.map(job => (
            <QueueItem 
              key={job.id} 
              {...convertJobToQueueItem(job)}
              onRemove={() => handleRemoveJob(job.id)}
            />
          ))
        ) : (
          <div className="text-center py-12 space-y-4">
            <div className="mx-auto w-16 h-16 bg-purple-500/20 rounded-full flex items-center justify-center">
              <ListIcon className="h-8 w-8 text-purple-400" />
            </div>
            <div className="space-y-2">
              <h3 className="text-lg font-medium text-gray-300">
                {searchQuery.trim() !== '' ? 'No downloads found' : 'No downloads in queue'}
              </h3>
              <p className="text-sm text-gray-400 max-w-md mx-auto">
                {searchQuery.trim() !== '' 
                  ? `No downloads match "${searchQuery}". Try adjusting your search or filter.`
                  : 'Start downloading by adding URLs in the Download tab. Active downloads will appear here, and completed downloads will stay in the "Completed" tab until History functionality is implemented.'
                }
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default QueueManager 