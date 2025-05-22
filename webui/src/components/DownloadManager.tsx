import React, { useState, useRef, useEffect, useCallback } from 'react'
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Textarea } from "@/components/ui/textarea"
import { Label } from "@/components/ui/label"

import { toast } from 'sonner'
import { type AppConfig, type DownloadOptions, type AddDownloadResponseItem } from "@/types/api"
import { 
  ClipboardPaste, 
  Folder, 
  Download,
  Loader2,
  Check
} from "lucide-react"

interface DownloadManagerProps {
  onStartDownload?: () => void
}

const DownloadManager = ({ onStartDownload }: DownloadManagerProps) => {
  const [isLoading, setIsLoading] = useState(false)
  const [url, setUrl] = useState('')
  const [downloadPath, setDownloadPath] = useState('')
  const [isUpdatingPath, setIsUpdatingPath] = useState(false)
  const [pathUpdateSuccess, setPathUpdateSuccess] = useState(false)
  const startButtonRef = useRef<HTMLButtonElement>(null)
  const pathUpdateTimeoutRef = useRef<NodeJS.Timeout | null>(null)

  // Function to update the config with new download path
  const updateConfigPath = useCallback(async (newPath: string) => {
    if (!newPath.trim()) return // Don't update if empty
    
    setIsUpdatingPath(true)
    try {
      // First get current config
      const response = await fetch('/api/config')
      if (!response.ok) {
        throw new Error('Failed to load current configuration')
      }
      
      const currentConfig: AppConfig = await response.json()
      
      // Update only the outPath
      const updatedConfig: AppConfig = {
        ...currentConfig,
        outPath: newPath.trim()
      }
      
      // Save updated config
      const saveResponse = await fetch('/api/config', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(updatedConfig),
      })
      
      if (!saveResponse.ok) {
        throw new Error('Failed to save configuration')
      }
      
      setPathUpdateSuccess(true)
      setTimeout(() => setPathUpdateSuccess(false), 2000) // Show success for 2 seconds
      
    } catch (error) {
      console.error('Error updating download path:', error)
      toast.error('Failed to update download path', {
        description: error instanceof Error ? error.message : 'Unknown error occurred'
      })
    } finally {
      setIsUpdatingPath(false)
    }
  }, [])

  // Debounced effect to update config when downloadPath changes
  useEffect(() => {
    if (pathUpdateTimeoutRef.current) {
      clearTimeout(pathUpdateTimeoutRef.current)
    }
    
    if (downloadPath.trim()) {
      pathUpdateTimeoutRef.current = setTimeout(() => {
        updateConfigPath(downloadPath)
      }, 1000) // Wait 1 second after user stops typing
    }
    
    return () => {
      if (pathUpdateTimeoutRef.current) {
        clearTimeout(pathUpdateTimeoutRef.current)
      }
    }
  }, [downloadPath, updateConfigPath])

  const handlePasteUrl = async () => {
    try {
      const text = await navigator.clipboard.readText()
      setUrl(text)
      toast.success('URL pasted successfully')
      
      // Set focus to start button if URL is valid
      if (text.trim() && startButtonRef.current) {
        startButtonRef.current.focus()
      }
    } catch (err) {
      console.error('Failed to read clipboard contents: ', err)
      toast.error('Failed to read clipboard', {
        description: 'Please paste the URL manually',
      })
    }
  }

  const handleSelectFolder = () => {
    toast.info('Folder selection', {
      description: 'Folder selection will be implemented in a future update',
    })
  }

  const handleStartDownload = async () => {
    if (!url.trim()) {
      toast.error('Missing URL', {
        description: 'Please enter a nugs.net URL to download',
      })
      return
    }
    
    setIsLoading(true)
    
    // Parse URLs from textarea (supports multiple URLs, one per line)
    const urlList = url.split(/\r?\n/).map(url => url.trim()).filter(url => url !== '')
    
    if (urlList.length === 0) {
      toast.error('No valid URLs found', {
        description: 'Please enter at least one valid URL',
      })
      setIsLoading(false)
      return
    }

    // Prepare download options (using defaults for now)
    const options: DownloadOptions = {
      forceVideo: false,
      skipVideos: false,
      skipChapters: false,
    }

    const payload = {
      urls: urlList,
      options: options,
    }

    console.log('Submitting download request:', payload)
    toast.info('Adding job(s) to queue...')
    
    try {
      const response = await fetch('/api/downloads', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(payload),
      })
      
      if (!response.ok) {
        let errorMsg = `HTTP error! status: ${response.status}`
        try {
          const errorData = await response.json()
          errorMsg = errorData.error || errorMsg
        } catch {
          errorMsg = response.statusText || errorMsg
        }
        throw new Error(errorMsg)
      }
      
      const results = await response.json()
      
      // Count successful jobs
      const successfulJobs = results.filter((result: AddDownloadResponseItem) => result.jobId && !result.error)
      const failedJobs = results.filter((result: AddDownloadResponseItem) => result.error)
      
      if (successfulJobs.length > 0) {
        toast.success(`${successfulJobs.length} download(s) added to queue!`, {
          description: successfulJobs.length === 1 
            ? `Job ID: ${successfulJobs[0].jobId.substring(0, 8)}...`
            : `${successfulJobs.length} downloads started`
        })
        setUrl('') // Clear the input on success
        
        // Switch to queue tab to show the newly added downloads
        if (onStartDownload) {
          onStartDownload()
        }
      }
      
      if (failedJobs.length > 0) {
        toast.error(`${failedJobs.length} URL(s) failed`, {
          description: failedJobs[0].error || 'Check console for details'
        })
      }
      
    } catch (err: unknown) {
      console.error('Error submitting download request:', err)
      const errorMessage = err instanceof Error ? err.message : 'Unknown error occurred'
      toast.error('Failed to start download', {
        description: errorMessage
      })
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-2">
        <Download className="h-5 w-5 text-purple-400" />
        <h2 className="text-lg font-medium">Download Manager</h2>
      </div>
      
      {/* URL Input Section */}
      <div className="space-y-3">
        <Label htmlFor="url-input" className="text-sm font-medium">
          Enter URLs to Download
        </Label>
        <Textarea 
          id="url-input"
          className="min-h-[160px] font-mono text-sm resize-y bg-gray-900/50 border-gray-600 focus:border-purple-400 focus:ring-1 focus:ring-purple-400"
          placeholder="Enter URLs to download (one per line)
Example: https://play.nugs.net/release/23329"
          value={url}
          onChange={(e) => setUrl(e.target.value)}
        />
        <div className="flex gap-2">
          <Button 
            variant="outline" 
            size="sm" 
            onClick={handlePasteUrl}
            className="flex items-center gap-1.5 text-xs border-gray-600 hover:border-gray-500"
          >
            <ClipboardPaste className="h-3.5 w-3.5" />
            <span>Paste URL</span>
          </Button>
        </div>
      </div>

      {/* Download Location Section */}
      <div className="space-y-3">
        <Label htmlFor="download-path" className="text-sm font-medium">
          Custom Download Location <span className="text-gray-400 text-sm">(Optional)</span>
        </Label>
        <div className="flex gap-2">
          <div className="relative flex-1">
            <Input 
              id="download-path"
              className="text-sm bg-gray-900/50 border-gray-600 focus:border-purple-400 focus:ring-1 focus:ring-purple-400 pr-8"
              placeholder="Leave empty to use default location from settings"
              value={downloadPath}
              onChange={(e) => setDownloadPath(e.target.value)}
            />
            {/* Status indicator */}
            <div className="absolute right-2 top-1/2 -translate-y-1/2">
              {isUpdatingPath && (
                <Loader2 className="h-4 w-4 animate-spin text-gray-400" />
              )}
              {pathUpdateSuccess && (
                <Check className="h-4 w-4 text-green-400" />
              )}
            </div>
          </div>
          <Button 
            variant="outline" 
            size="icon" 
            onClick={handleSelectFolder}
            className="shrink-0 border-gray-600 hover:border-gray-500"
          >
            <Folder className="h-4 w-4" />
          </Button>
        </div>
        <p className="text-xs text-gray-400">
          {downloadPath.trim() ? (
            <span className="text-purple-300">
              This path will override your default configuration. Changes are saved automatically.
            </span>
          ) : (
            "If specified, this will override the default download location in your configuration."
          )}
        </p>
      </div>



      {/* Start Download Button */}
      <Button 
        className="w-full h-12 bg-gradient-to-r from-purple-600 to-pink-600 hover:from-purple-700 hover:to-pink-700 text-white font-medium text-sm"
        disabled={isLoading || !url.trim()}
        onClick={handleStartDownload}
        ref={startButtonRef}
      >
        {isLoading ? (
          <>
            <Loader2 className="mr-2 h-4 w-4 animate-spin" />
            <span>Processing...</span>
          </>
        ) : (
          <>
            <Download className="mr-2 h-4 w-4" />
            <span>Start Download</span>
          </>
        )}
      </Button>
    </div>
  )
}

export default DownloadManager 