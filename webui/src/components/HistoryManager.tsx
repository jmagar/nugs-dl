import { useState, useEffect, useCallback } from 'react'; // React import removed, useCallback added
import { Input } from "@/components/ui/input"
import { Button } from "@/components/ui/button"
import { Badge } from "@/components/ui/badge"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { toast } from 'sonner'
import { type DownloadJob } from "@/types/api"
import { 
  Search, 
  RefreshCw, 
  Trash2,
  Music,
  PlayCircle,
  Calendar,
  ListMusic,
  FolderOpen,
  ExternalLink,
  HistoryIcon,
  Loader2,
  Download,
  Share
} from "lucide-react"

// Convert DownloadJob to local history item interface
interface HistoryItemProps {
  id: string
  title: string
  type: 'album' | 'video' | 'livestream' | 'playlist'
  url: string
  path: string
  date: string
  size: string
  format: string
}

const HistoryItem = ({ item }: { item: HistoryItemProps }) => {
  
  // Return appropriate icon and badge styling based on content type
  const getTypeBadge = () => {
    switch (item.type) {
      case 'album':
        return (
          <Badge className="bg-purple-500/20 text-purple-400 border-purple-500/30 text-xs">
            <Music className="h-3 w-3 mr-1" />
            album
          </Badge>
        )
      case 'video':
        return (
          <Badge className="bg-blue-500/20 text-blue-400 border-blue-500/30 text-xs">
            <PlayCircle className="h-3 w-3 mr-1" />
            video
          </Badge>
        )
      case 'livestream':
        return (
          <Badge className="bg-green-500/20 text-green-400 border-green-500/30 text-xs">
            <Calendar className="h-3 w-3 mr-1" />
            livestream
          </Badge>
        )
      case 'playlist':
        return (
          <Badge className="bg-yellow-500/20 text-yellow-400 border-yellow-500/30 text-xs">
            <ListMusic className="h-3 w-3 mr-1" />
            playlist
          </Badge>
        )
      default:
        return null
    }
  }

  const handleOpenUrl = () => {
    window.open(item.url, '_blank')
  }

  const handleDownload = async () => {
    try {
      // Check if download is available
      const downloadUrl = `/api/download/${item.id}`
      const response = await fetch(downloadUrl, { method: 'HEAD' })
      
      if (response.ok) {
        // Create a temporary link and trigger download
        const link = document.createElement('a')
        link.href = downloadUrl
        link.download = `${item.title}.zip` // Suggest a filename
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        
        toast.success('Download started!')
      } else if (response.status === 501) {
        toast.info('File download feature is coming soon!', {
          description: 'The backend is still being implemented for file serving.'
        })
      } else {
        throw new Error(`Server returned ${response.status}`)
      }
    } catch (err) {
      console.error('Error starting download:', err)
      toast.error('Failed to start download', {
        description: 'Please check if the file is still available on the server.'
      })
    }
  }

  const handleShare = async () => {
    try {
      const downloadUrl = `${window.location.origin}/api/download/${item.id}`
      
      // Try to copy to clipboard
      await navigator.clipboard.writeText(downloadUrl)
      toast.success('Download link copied to clipboard!')
    } catch (err) {
      console.error('Error copying to clipboard:', err)
      // Fallback: show the URL in a prompt
      window.prompt('Download link:', `${window.location.origin}/api/download/${item.id}`)
      toast.info('Download link ready to copy')
    }
  }

  const handleDelete = async () => {
    if (window.confirm(`Are you sure you want to delete "${item.title}" from your history?`)) {
      try {
        // TODO: Implement actual deletion when backend endpoint exists
        toast.info('History deletion not yet implemented')
      } catch (err) {
        console.error('Error deleting history item:', err)
        toast.error('Failed to delete history item')
      }
    }
  }

  return (
    <div className="bg-gray-800/30 border border-gray-700 rounded-lg p-4">
      <div className="flex flex-col gap-3">
        {/* Title row with badge */}
        <div className="flex items-center justify-between mb-3">
          <div className="flex items-center gap-2 flex-1 min-w-0">
            <h3 className="text-sm font-medium text-white truncate">{item.title}</h3>
            {getTypeBadge()}
          </div>
          <div className="flex gap-2 ml-4">
            <Button 
              variant="outline" 
              size="sm" 
              className="border-gray-600 hover:border-gray-500"
              onClick={handleOpenUrl}
              aria-label={`Visit URL for ${item.title}`}
            >
              <ExternalLink className="h-3.5 w-3.5 mr-1" />
              Visit
            </Button>
            <Button 
              variant="outline" 
              size="sm" 
              className="border-gray-600 hover:border-gray-500"
              onClick={handleDownload}
              aria-label={`Download ${item.title}`}
            >
              <Download className="h-3.5 w-3.5 mr-1" />
              Download
            </Button>
            <Button 
              variant="outline" 
              size="sm" 
              className="border-gray-600 hover:border-gray-500"
              onClick={handleShare}
              aria-label={`Share download link for ${item.title}`}
            >
              <Share className="h-3.5 w-3.5 mr-1" />
              Share
            </Button>
            <Button 
              variant="outline" 
              size="sm" 
              className="border-gray-600 hover:border-gray-500 text-red-400 hover:text-red-300"
              onClick={handleDelete}
              aria-label={`Delete ${item.title} from history`}
            >
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          </div>
        </div>
        
        {/* URL */}
        <div className="mb-2 text-xs text-blue-400 font-mono break-all">
          {item.url}
        </div>
        
        {/* Path */}
        <div className="flex items-center gap-2 mb-3 text-xs text-gray-300">
          <FolderOpen className="h-3 w-3 text-gray-400" />
          <span className="truncate">{item.path}</span>
        </div>
        
        {/* Metadata */}
        <div className="flex items-center gap-4 text-xs text-gray-400">
          <span>{item.date}</span>
          <span>{item.size}</span>
          <span>{item.format}</span>
        </div>
      </div>
    </div>
  )
}

const HistoryManager = () => {
  const [searchQuery, setSearchQuery] = useState('')
  const [typeFilter, setTypeFilter] = useState('all')
  const [historyItems, setHistoryItems] = useState<HistoryItemProps[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  // Extract title from URL (fallback for when job.title is not set)
  const extractTitleFromUrl = useCallback((url: string): string => {
    try {
      const urlObj = new URL(url)
      const pathParts = urlObj.pathname.split('/').filter(part => part.length > 0)
      
      if (urlObj.hostname.includes('nugs.net')) {
        if (pathParts.includes('release') && pathParts.length >= 2) {
          const releaseId = pathParts[pathParts.indexOf('release') + 1]
          return `Release ${releaseId}`
        }
        if (pathParts.includes('artist') && pathParts.length >= 2) {
          const artistId = pathParts[pathParts.indexOf('artist') + 1]
          return `Artist ${artistId}`
        }
        // Add other nugs.net specific patterns if needed
      }
      
      // Fallback if no specific pattern matches or not a nugs.net URL
      const lastPart = pathParts[pathParts.length - 1]
      if (lastPart && lastPart.length > 0) {
        // If last part is numeric, assume it's an ID-like title part
        if (/^\d+$/.test(lastPart)) {
          return `Download ${lastPart}` 
        }
        // Otherwise, format it as a title
        return lastPart.replace(/-/g, ' ').replace(/\b\w/g, l => l.toUpperCase());
      }
      
      return 'Unknown Title' // Ultimate fallback
    } catch (e) {
      console.warn(`Error parsing URL for title: ${url}`, e)
      return 'Invalid URL' // Fallback for parsing errors
    }
  }, [])

  // Helper function to convert DownloadJob to HistoryItemProps
  const convertJobToHistoryItem = useCallback((job: DownloadJob): HistoryItemProps => {
    // Extract title from job title or URL
    const title = job.title || extractTitleFromUrl(job.originalUrl)
    
    return {
      id: job.id,
      title: title,
      type: 'album', // Default to album, could be enhanced based on URL analysis
      url: job.originalUrl,
      path: 'N/A', // Backend doesn't provide download path yet
      date: job.completedAt ? new Date(job.completedAt).toLocaleDateString() : 'Unknown',
      size: 'Unknown', // Backend doesn't provide file size yet
      format: 'FLAC' // Default format
    }
  }, [extractTitleFromUrl])
  
  // Load history data
  useEffect(() => {
    const fetchHistory = async () => {
      setIsLoading(true)
      setError(null)
      try {
        const response = await fetch('/api/history')
        if (!response.ok) {
          throw new Error(`Failed to fetch history: ${response.statusText}`)
        }
        const completedJobs: DownloadJob[] = await response.json()
        
        // Ensure we have a valid array (defensive programming)
        if (!Array.isArray(completedJobs)) {
          console.error('History API returned non-array:', completedJobs)
          throw new Error('Invalid response format from history API')
        }
        
        // Convert completed jobs to history items
        const items = completedJobs.map(convertJobToHistoryItem)
        setHistoryItems(items)
      } catch (err: unknown) {
        console.error('Error fetching history:', err)
        const errorMessage = err instanceof Error ? err.message : 'Failed to load download history'
        setError(errorMessage)
        toast.error(errorMessage)
      } finally {
        setIsLoading(false)
      }
    }

    fetchHistory()
  }, [convertJobToHistoryItem])
  
  const handleRefresh = async () => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await fetch('/api/history')
      if (!response.ok) {
        throw new Error(`Failed to refresh history: ${response.statusText}`)
      }
      const completedJobs: DownloadJob[] = await response.json()
      
      // Convert completed jobs to history items
      const items = completedJobs.map(convertJobToHistoryItem)
      setHistoryItems(items)
      toast.success('History refreshed successfully')
    } catch (err: unknown) {
      console.error('Error refreshing history:', err)
      const errorMessage = err instanceof Error ? err.message : 'Failed to refresh download history'
      setError(errorMessage)
      toast.error(errorMessage)
    } finally {
      setIsLoading(false)
    }
  }
  
  const handleClearHistory = () => {
    if (window.confirm('Are you sure you want to clear all download history?')) {
      toast.info('Clear history not yet implemented - waiting for backend endpoint')
    }
  }
  
  // Filter items based on type filter and search query
  const filteredItems = historyItems.filter(item => {
    // First filter by type
    if (typeFilter !== 'all' && item.type !== typeFilter) {
      return false
    }
    
    // Then filter by search query if present
    if (searchQuery.trim() !== '') {
      return (
        item.title.toLowerCase().includes(searchQuery.toLowerCase()) ||
        item.url.toLowerCase().includes(searchQuery.toLowerCase()) ||
        item.path.toLowerCase().includes(searchQuery.toLowerCase())
      )
    }
    
    return true
  })

  if (isLoading) {
    return (
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <HistoryIcon className="h-5 w-5 text-purple-400" />
          <h2 className="text-lg font-medium">Download History</h2>
        </div>
        <div className="flex items-center justify-center py-12">
          <Loader2 className="h-8 w-8 animate-spin text-purple-400" />
          <span className="ml-2 text-gray-300">Loading history...</span>
        </div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="space-y-4">
        <div className="flex items-center gap-2">
          <HistoryIcon className="h-5 w-5 text-purple-400" />
          <h2 className="text-lg font-medium">Download History</h2>
        </div>
        <div className="text-center py-12 space-y-4">
          <div className="mx-auto w-16 h-16 bg-red-500/20 rounded-full flex items-center justify-center">
            <HistoryIcon className="h-8 w-8 text-red-400" />
          </div>
          <div className="space-y-2">
            <h3 className="text-lg font-medium text-gray-300">Failed to load history</h3>
            <p className="text-sm text-gray-400 max-w-md mx-auto">{error}</p>
          </div>
        </div>
      </div>
    )
  }

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <HistoryIcon className="h-5 w-5 text-purple-400" />
        <h2 className="text-lg font-medium">Download History</h2>
      </div>

      {/* Search Bar */}
      <div className="relative">
        <Search className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-gray-400" />
        <Input
          placeholder="Search history..."
          value={searchQuery}
          onChange={(e) => setSearchQuery(e.target.value)}
          className="pl-10 bg-gray-900/50 border-gray-600 focus:border-purple-400 focus:ring-1 focus:ring-purple-400"
          aria-label="Search download history"
        />
      </div>

      {/* Action Buttons */}
      <div className="flex gap-2">
        <Select value={typeFilter} onValueChange={setTypeFilter}>
          <SelectTrigger className="w-[160px] bg-gray-900/50 border-gray-600">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All Types</SelectItem>
            <SelectItem value="album">Albums</SelectItem>
            <SelectItem value="video">Videos</SelectItem>
            <SelectItem value="livestream">Livestreams</SelectItem>
            <SelectItem value="playlist">Playlists</SelectItem>
          </SelectContent>
        </Select>
        
        <Button 
          variant="outline" 
          size="sm" 
          className="flex items-center gap-1.5 border-gray-600 hover:border-gray-500"
          onClick={handleRefresh}
          aria-label="Refresh history list"
        >
          <RefreshCw className="h-4 w-4" />
          Refresh
        </Button>
        <Button 
          variant="outline" 
          size="sm" 
          className="flex items-center gap-1.5 border-gray-600 hover:border-gray-500 text-red-400 hover:text-red-300"
          onClick={handleClearHistory}
          aria-label="Clear all download history"
        >
          <Trash2 className="h-4 w-4" />
          Clear All
        </Button>
      </div>

      {/* History Items */}
      <div className="space-y-3">
        {filteredItems.length > 0 ? (
          <>
            <div className="sr-only" aria-live="polite">
              {`Showing ${filteredItems.length} ${typeFilter === 'all' ? 'items' : typeFilter + 's'} ${searchQuery ? `matching "${searchQuery}"` : ''}`}
            </div>
            {filteredItems.map(item => (
              <HistoryItem key={item.id} item={item} />
            ))}
          </>
        ) : (
          <div className="text-center py-12 space-y-4">
            <div className="mx-auto w-16 h-16 bg-purple-500/20 rounded-full flex items-center justify-center">
              <HistoryIcon className="h-8 w-8 text-purple-400" />
            </div>
            <div className="space-y-2">
              <h3 className="text-lg font-medium text-gray-300">
                No download history
              </h3>
              <p className="text-sm text-gray-400 max-w-md mx-auto">
                {searchQuery.trim() !== '' || typeFilter !== 'all'
                  ? 'No downloads match the current filter. Try adjusting your search or type filter.'
                  : 'No completed downloads yet. Completed downloads will appear here with download information and metadata.'
                }
              </p>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}

export default HistoryManager 