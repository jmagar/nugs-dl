import React from 'react'
import { Badge } from "@/components/ui/badge"
import { Button } from "@/components/ui/button"
import { Progress } from "@/components/ui/progress"
import { 
  Pause, 
  Play, 
  X, 
  Music, 
  PlayCircle,
  RefreshCw
} from "lucide-react"

export interface QueueItemProps {
  id: string
  title: string
  type: 'album' | 'video' | 'livestream' | 'playlist'
  status: 'downloading' | 'queued' | 'paused' | 'completed' | 'error'
  progress: number
  format: string
  size: string
  speed?: string
  eta?: string
  errorMessage?: string
  currentTrack?: number
  totalTracks?: number
  artworkUrl?: string
  onRemove?: () => void
}

const QueueItem = ({ 
  title, 
  type, 
  status, 
  progress, 
  format, 
  size, 
  speed, 
  eta,
  currentTrack,
  totalTracks,
  artworkUrl,
  onRemove
}: QueueItemProps) => {
  
  // Return appropriate action buttons based on status
  const renderActionButtons = () => {
    switch (status) {
      case 'downloading':
        return (
          <Button 
            variant="outline" 
            size="sm" 
            className="border-gray-600 hover:border-gray-500"
          >
            <Pause className="h-3.5 w-3.5" />
            Pause
          </Button>
        )
      case 'queued':
        return (
          <>
            <Button 
              variant="outline" 
              size="sm" 
              className="border-gray-600 hover:border-gray-500"
            >
              <Play className="h-3.5 w-3.5" />
              Start
            </Button>
            <Button 
              variant="outline" 
              size="sm" 
              className="border-gray-600 hover:border-gray-500 text-red-400 hover:text-red-300"
              onClick={onRemove}
            >
              <X className="h-3.5 w-3.5" />
              Cancel
            </Button>
          </>
        )
      case 'paused':
        return (
          <>
            <Button 
              variant="outline" 
              size="sm" 
              className="border-gray-600 hover:border-gray-500"
            >
              <Play className="h-3.5 w-3.5" />
              Resume
            </Button>
            <Button 
              variant="outline" 
              size="sm" 
              className="border-gray-600 hover:border-gray-500 text-red-400 hover:text-red-300"
              onClick={onRemove}
            >
              <X className="h-3.5 w-3.5" />
              Cancel
            </Button>
          </>
        )
      case 'error':
        return (
          <>
            <Button 
              variant="outline" 
              size="sm" 
              className="border-gray-600 hover:border-gray-500"
            >
              <RefreshCw className="h-3.5 w-3.5" />
              Retry
            </Button>
            <Button 
              variant="outline" 
              size="sm" 
              className="border-gray-600 hover:border-gray-500 text-red-400 hover:text-red-300"
              onClick={onRemove}
            >
              <X className="h-3.5 w-3.5" />
              Cancel
            </Button>
          </>
        )
      case 'completed':
        return null
      default:
        return null
    }
  }

  // Return appropriate icon based on content type
  const renderTypeIcon = () => {
    switch (type) {
      case 'album':
        return <Music className="h-4 w-4 text-purple-400" />
      case 'video':
      case 'livestream':
      case 'playlist':
        return <PlayCircle className="h-4 w-4 text-blue-400" />
      default:
        return null
    }
  }

  // Progress color with gradients based on status
  const getProgressColor = () => {
    switch (status) {
      case 'downloading': return 'bg-gradient-to-r from-green-500 to-emerald-400'
      case 'queued': return 'bg-gray-500'
      case 'paused': return 'bg-yellow-500'
      case 'completed': return 'bg-gradient-to-r from-green-500 to-emerald-400'
      case 'error': return 'bg-gradient-to-r from-red-500 to-pink-500'
      default: return 'bg-gradient-to-r from-green-500 to-emerald-400'
    }
  }

  // Enhanced status badge with animations
  const renderStatusBadge = () => {
    switch (status) {
      case 'downloading':
        return (
          <Badge className="bg-blue-500/20 text-blue-400 border-blue-500/30 text-xs animate-pulse">
            <div className="w-2 h-2 bg-blue-400 rounded-full mr-1.5 animate-ping"></div>
            downloading
          </Badge>
        )
      case 'queued':
        return (
          <Badge className="bg-gray-500/20 text-gray-400 border-gray-500/30 text-xs">
            <div className="w-2 h-2 bg-gray-400 rounded-full mr-1.5"></div>
            queued
          </Badge>
        )
      case 'paused':
        return (
          <Badge className="bg-yellow-500/20 text-yellow-400 border-yellow-500/30 text-xs">
            <div className="w-2 h-2 bg-yellow-400 rounded-full mr-1.5"></div>
            paused
          </Badge>
        )
      case 'completed':
        return (
          <Badge className="bg-green-500/20 text-green-400 border-green-500/30 text-xs">
            <div className="w-2 h-2 bg-green-400 rounded-full mr-1.5"></div>
            completed
          </Badge>
        )
      case 'error':
        return (
          <Badge className="bg-red-500/20 text-red-400 border-red-500/30 text-xs">
            <div className="w-2 h-2 bg-red-400 rounded-full mr-1.5 animate-pulse"></div>
            error
          </Badge>
        )
      default:
        return (
          <Badge className="bg-blue-500/20 text-blue-400 border-blue-500/30 text-xs">
            {status}
          </Badge>
        )
    }
  }

  return (
    <div className="bg-gray-800/30 border border-gray-700 rounded-lg p-4 hover:bg-gray-800/50 hover:border-gray-600 transition-all duration-300 animate-in slide-in-from-left-1">
      {/* Title and badges */}
      <div className="flex items-start justify-between mb-3">
        <div className="flex items-center gap-3 flex-1 min-w-0">
          {/* Artwork */}
          {artworkUrl ? (
            <img 
              src={artworkUrl} 
              alt="Album artwork" 
              className="w-12 h-12 rounded-md object-cover flex-shrink-0 shadow-lg border border-gray-600"
              onError={(e) => {
                e.currentTarget.style.display = 'none'
              }}
            />
          ) : (
            <div className="w-12 h-12 rounded-md bg-gray-700 flex items-center justify-center flex-shrink-0 border border-gray-600">
              {renderTypeIcon()}
            </div>
          )}
          
          <div className="flex items-center gap-2 flex-1 min-w-0">
            <h3 className="text-sm font-medium text-white truncate">{title}</h3>
            {renderStatusBadge()}
          </div>
        </div>
        <div className="flex gap-2 ml-4">
          {renderActionButtons()}
        </div>
      </div>
      
      {/* Progress bar */}
      <div className="mb-3">
        {progress >= 0 ? (
          <Progress 
            value={progress} 
            className="h-2 bg-gray-700"
            indicatorClassName={getProgressColor()}
          />
        ) : (
          // Indeterminate progress bar for unknown size
          <div className="h-2 bg-gray-700 rounded-full overflow-hidden">
            <div className="h-full w-full bg-gradient-to-r from-purple-500 to-pink-500 animate-pulse"></div>
          </div>
        )}
      </div>
      
      {/* Status info */}
      <div className="flex items-center gap-4 text-xs text-gray-300">
        <span>
          {progress >= 0 ? `${progress}%` : 'Processing...'}
          {currentTrack && totalTracks && ` (Track ${currentTrack}/${totalTracks})`}
        </span>
        {speed && <span>{speed}</span>}
        <span>{size}</span>
        <span>{format}</span>
        {eta && <span>ETA: {eta}</span>}
      </div>
    </div>
  )
}

export default QueueItem 