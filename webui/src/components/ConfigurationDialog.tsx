import React, { useState, useEffect } from 'react'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
  DialogClose
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select"
import { Loader2, Folder, Save } from "lucide-react"
import { toast } from "sonner"
import { type AppConfig } from "@/types/api"

interface ConfigurationDialogProps {
  open: boolean
  onOpenChange: (open: boolean) => void
}

// Format mappings to match the backend config
const audioFormatOptions = [
  { value: '2', label: '2 - 16-bit / 44.1 kHz FLAC' },
  { value: '1', label: '1 - ALAC (16-bit/44.1kHz)' },
  { value: '3', label: '3 - MQA (24-bit/48kHz)' },
  { value: '4', label: '4 - 360/Best Available' },
  { value: '5', label: '5 - AAC (150kbps)' },
]

const videoFormatOptions = [
  { value: '3', label: '3 - 1080p' },
  { value: '1', label: '1 - 480p' },
  { value: '2', label: '2 - 720p' },
  { value: '4', label: '4 - 1440p' },
  { value: '5', label: '5 - 4K/Best Available' },
]

const ConfigurationDialog = ({ open, onOpenChange }: ConfigurationDialogProps) => {
  const [isLoading, setIsLoading] = useState(false)
  const [isInitialLoading, setIsInitialLoading] = useState(false)
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [audioFormat, setAudioFormat] = useState('2') // Default to FLAC
  const [videoFormat, setVideoFormat] = useState('3') // Default to 1080p
  const [downloadPath, setDownloadPath] = useState('')
  const [authToken, setAuthToken] = useState('')
  const [useFFmpeg, setUseFFmpeg] = useState(false)
  const [formErrors, setFormErrors] = useState<Record<string, string>>({})

  // Load configuration when dialog opens
  useEffect(() => {
    if (open) {
      loadConfiguration()
    }
  }, [open])

  const loadConfiguration = async () => {
    setIsInitialLoading(true)
    try {
      const response = await fetch('/api/config')
      if (!response.ok) {
        throw new Error(`Failed to load configuration: ${response.statusText}`)
      }
      
      const config: AppConfig = await response.json()
      
      // Set form values from loaded config
      setEmail(config.email || '')
      setPassword(config.password || '')
      setAudioFormat(config.format?.toString() || '2')
      setVideoFormat(config.videoFormat?.toString() || '3')
      setDownloadPath(config.outPath || '')
      setAuthToken(config.token || '')
      setUseFFmpeg(config.useFfmpegEnvVar || false)
      
    } catch (error) {
      console.error('Error loading configuration:', error)
      toast.error('Failed to load configuration', {
        description: 'Using default values'
      })
    } finally {
      setIsInitialLoading(false)
    }
  }

  const handleSelectFolder = () => {
    toast.info('Folder selection', {
      description: 'Folder selection will be implemented in a future update',
    })
  }

  const validateForm = () => {
    const errors: Record<string, string> = {};
    
    if (!email.trim()) {
      errors.email = 'Email is required';
    } else if (!/^\S+@\S+\.\S+$/.test(email)) {
      errors.email = 'Email format is invalid';
    }
    
    if (!password.trim()) {
      errors.password = 'Password is required';
    }
    
    setFormErrors(errors);
    return Object.keys(errors).length === 0;
  };

  const handleSave = async () => {
    if (!validateForm()) {
      toast.error('Validation failed', {
        description: 'Please check the form for errors',
      })
      return
    }

    setIsLoading(true)
    
    try {
      const configData: AppConfig = {
        email: email.trim(),
        password: password.trim(),
        format: parseInt(audioFormat),
        videoFormat: parseInt(videoFormat),
        outPath: downloadPath.trim() || 'downloads',
        token: authToken.trim(),
        useFfmpegEnvVar: useFFmpeg
      }

      const response = await fetch('/api/config', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(configData),
      })

      if (!response.ok) {
        const errorData = await response.text()
        throw new Error(errorData || `Failed to save configuration: ${response.statusText}`)
      }

      onOpenChange(false)
      toast.success('Configuration saved', {
        description: 'Your settings have been updated successfully',
      })
    } catch (error) {
      console.error('Error saving configuration:', error)
      toast.error('Failed to save configuration', {
        description: error instanceof Error ? error.message : 'Unknown error occurred'
      })
    } finally {
      setIsLoading(false)
    }
  }

  if (isInitialLoading) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent className="bg-gray-800/95 backdrop-blur-sm border border-gray-700 shadow-xl rounded-lg w-full max-w-md p-6 max-h-[95vh] overflow-y-auto [&>button]:text-gray-400 [&>button]:hover:text-white">
          <div className="flex items-center justify-center py-8">
            <Loader2 className="h-8 w-8 animate-spin text-purple-400" />
            <span className="ml-2 text-gray-300">Loading configuration...</span>
          </div>
        </DialogContent>
      </Dialog>
    )
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="bg-gray-800/95 backdrop-blur-sm border border-gray-700 shadow-xl rounded-lg w-full max-w-md p-6 max-h-[95vh] overflow-y-auto [&>button]:text-gray-400 [&>button]:hover:text-white">
        <DialogHeader className="pb-4">
          <DialogTitle className="text-xl font-semibold text-white">Configuration Settings</DialogTitle>
          <p className="text-sm text-gray-300 mt-1">
            Configure your nugs.net credentials and download preferences.
          </p>
        </DialogHeader>
        
        <div className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="email" className="text-sm font-medium text-white">
              Email
            </Label>
            <Input
              id="email"
              type="email"
              placeholder="your@email.com"
              value={email}
              onChange={(e) => {
                setEmail(e.target.value)
                if (formErrors.email) {
                  setFormErrors({...formErrors, email: ''})
                }
              }}
              className={`bg-gray-900/50 border-gray-600 text-white placeholder:text-gray-400 focus:border-purple-400 focus:ring-1 focus:ring-purple-400 ${formErrors.email ? 'border-red-500' : ''}`}
              required
            />
            {formErrors.email && (
              <p className="text-xs text-red-400 mt-1">
                {formErrors.email}
              </p>
            )}
          </div>
          
          <div className="space-y-2">
            <Label htmlFor="password" className="text-sm font-medium text-white">
              Password
            </Label>
            <Input
              id="password"
              type="password"
              placeholder="••••••••"
              value={password}
              onChange={(e) => {
                setPassword(e.target.value)
                if (formErrors.password) {
                  setFormErrors({...formErrors, password: ''})
                }
              }}
              className={`bg-gray-900/50 border-gray-600 text-white placeholder:text-gray-400 focus:border-purple-400 focus:ring-1 focus:ring-purple-400 ${formErrors.password ? 'border-red-500' : ''}`}
              required
            />
            {formErrors.password && (
              <p className="text-xs text-red-400 mt-1">
                {formErrors.password}
              </p>
            )}
          </div>
          
          <div className="space-y-2">
            <Label htmlFor="audio-format" className="text-sm font-medium text-white">
              Audio Format
            </Label>
            <Select value={audioFormat} onValueChange={setAudioFormat}>
              <SelectTrigger id="audio-format" className="bg-gray-900/50 border-gray-600 text-white">
                <SelectValue placeholder="Select format" />
              </SelectTrigger>
              <SelectContent>
                {audioFormatOptions.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className="text-xs text-gray-400">Track download quality</p>
          </div>
          
          <div className="space-y-2">
            <Label htmlFor="video-format" className="text-sm font-medium text-white">
              Video Format
            </Label>
            <Select value={videoFormat} onValueChange={setVideoFormat}>
              <SelectTrigger id="video-format" className="bg-gray-900/50 border-gray-600 text-white">
                <SelectValue placeholder="Select format" />
              </SelectTrigger>
              <SelectContent>
                {videoFormatOptions.map((option) => (
                  <SelectItem key={option.value} value={option.value}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            <p className="text-xs text-gray-400">Video download format. FFmpeg needed.</p>
          </div>
          
          <div className="space-y-2">
            <Label htmlFor="download-path" className="text-sm font-medium text-white">
              Default Download Path
            </Label>
            <div className="flex gap-2">
              <Input
                id="download-path"
                placeholder="downloads"
                value={downloadPath}
                onChange={(e) => setDownloadPath(e.target.value)}
                className="flex-1 bg-gray-900/50 border-gray-600 text-white placeholder:text-gray-400 focus:border-purple-400 focus:ring-1 focus:ring-purple-400"
              />
              <Button 
                variant="outline" 
                size="icon" 
                onClick={handleSelectFolder}
                className="border-gray-600 hover:border-gray-500 hover:bg-gray-700"
              >
                <Folder className="h-4 w-4" />
              </Button>
            </div>
            <p className="text-xs text-gray-400">
              Path will be created if it doesn't exist.
            </p>
          </div>
          
          <div className="space-y-2">
            <Label htmlFor="auth-token" className="text-sm font-medium text-white">
              Auth Token <span className="text-gray-400">(Optional)</span>
            </Label>
            <Input
              id="auth-token"
              placeholder="For Apple/Google accounts"
              value={authToken}
              onChange={(e) => setAuthToken(e.target.value)}
              className="bg-gray-900/50 border-gray-600 text-white placeholder:text-gray-400 focus:border-purple-400 focus:ring-1 focus:ring-purple-400"
            />
            <p className="text-xs text-gray-400">
              Required only for Apple and Google accounts
            </p>
          </div>
          
          <div className="flex items-center justify-between py-2">
            <div>
              <Label htmlFor="use-ffmpeg" className="text-sm font-medium text-white">
                Use FFmpeg from environment variable
              </Label>
            </div>
            <Switch 
              id="use-ffmpeg" 
              checked={useFFmpeg}
              onCheckedChange={setUseFFmpeg}
            />
          </div>
        </div>
        
        <DialogFooter className="pt-6 flex gap-3">
          <DialogClose asChild>
            <Button 
              variant="outline" 
              className="border-gray-600 hover:border-gray-500 hover:bg-gray-700 text-white"
            >
              Cancel
            </Button>
          </DialogClose>
          <Button 
            onClick={handleSave}
            disabled={isLoading}
            className="bg-gradient-to-r from-purple-600 to-pink-600 hover:from-purple-700 hover:to-pink-700 text-white"
          >
            {isLoading ? (
              <>
                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                Saving...
              </>
            ) : (
              <>
                <Save className="mr-2 h-4 w-4" />
                Save Configuration
              </>
            )}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}

export default ConfigurationDialog 