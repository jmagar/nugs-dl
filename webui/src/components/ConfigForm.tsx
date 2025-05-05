import React, { useState } from 'react';
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Switch } from "@/components/ui/switch";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { type AppConfig } from "@/types/api";

// Format options
const audioFormatOptions = [
  { value: 1, label: '1: ALAC (16-bit/44.1kHz)' },
  { value: 2, label: '2: FLAC (16-bit/44.1kHz)' },
  { value: 3, label: '3: MQA (24-bit/48kHz)' },
  { value: 4, label: '4: 360/Best Available' },
  { value: 5, label: '5: AAC (150kbps)' },
];

const videoFormatOptions = [
  { value: 1, label: '1: 480p' },
  { value: 2, label: '2: 720p' },
  { value: 3, label: '3: 1080p' },
  { value: 4, label: '4: 1440p' },
  { value: 5, label: '5: 4K/Best Available' },
];

interface ConfigFormProps {
  onSubmit: (config: AppConfig) => Promise<void>; 
  initialConfig: AppConfig | null;
  isLoading?: boolean;
}

export function ConfigForm({ onSubmit, initialConfig, isLoading } : ConfigFormProps) {
  // State now initialized from props
  const [config, setConfig] = useState<AppConfig>(initialConfig || {
    email: '',
    password: '',
    format: 2, 
    videoFormat: 3, 
    outPath: '',
    token: '',
    useFfmpegEnvVar: false,
  });

  // --- Handlers use camelCase --- 
  const handleChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value, type } = event.target;
    const key = name as keyof AppConfig;
    if (key in config) { 
      setConfig((prevConfig: AppConfig) => ({
        ...prevConfig,
        [key]: type === 'number' ? parseInt(value, 10) || 0 : value,
      }));
    }
  };

  // Handlers for Select components
  const handleAudioFormatChange = (value: string) => {
     setConfig(prev => ({ ...prev, format: parseInt(value, 10) || 2 }));
  };
  const handleVideoFormatChange = (value: string) => {
     setConfig(prev => ({ ...prev, videoFormat: parseInt(value, 10) || 3 }));
  };

  const handleSwitchChange = (checked: boolean) => {
    setConfig((prevConfig: AppConfig) => ({
      ...prevConfig,
      useFfmpegEnvVar: checked,
    }));
  };

  // Internal submit handler calls the prop
  const handleFormSubmit = (event: React.FormEvent) => {
    event.preventDefault();
    onSubmit(config); 
  };

  return (
    <form id="config-form" onSubmit={handleFormSubmit} className="space-y-4">
        {/* Email / Password */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="email">Email</Label>
            <Input id="email" name="email" type="email" placeholder="your@email.com" value={config.email} onChange={handleChange} disabled={isLoading} />
          </div>
          <div className="space-y-2">
            <Label htmlFor="password">Password</Label>
            <Input id="password" name="password" type="password" placeholder="••••••••" value={config.password} onChange={handleChange} disabled={isLoading} />
          </div>
        </div>
        {/* Token */}
        <div className="space-y-2">
          <Label htmlFor="token">Token (Optional)</Label>
          <Input id="token" name="token" type="password" placeholder="Enter token if not using email/password" value={config.token} onChange={handleChange} disabled={isLoading} />
          <p className="text-xs text-muted-foreground">Use this for Apple/Google logins. Leave password blank if using token.</p>
        </div>
        {/* Formats using Select */}
        <div className="grid grid-cols-1 sm:grid-cols-2 gap-4">
          <div className="space-y-2">
            <Label htmlFor="format">Audio Format</Label>
            <Select 
                value={String(config.format)} 
                onValueChange={handleAudioFormatChange}
                disabled={isLoading}
            >
              <SelectTrigger id="format">
                <SelectValue placeholder="Select audio format..." />
              </SelectTrigger>
              <SelectContent>
                {audioFormatOptions.map(option => (
                  <SelectItem key={option.value} value={String(option.value)}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <div className="space-y-2">
            <Label htmlFor="videoFormat">Video Format</Label>
             <Select 
                value={String(config.videoFormat)} 
                onValueChange={handleVideoFormatChange}
                disabled={isLoading}
            >
              <SelectTrigger id="videoFormat">
                <SelectValue placeholder="Select video format..." />
              </SelectTrigger>
              <SelectContent>
                {videoFormatOptions.map(option => (
                  <SelectItem key={option.value} value={String(option.value)}>
                    {option.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
        </div>
        {/* Output Path */}
        <div className="space-y-2">
          <Label htmlFor="outPath">Output Path</Label>
          <Input id="outPath" name="outPath" type="text" placeholder="e.g., /Users/you/Downloads/Nugs" value={config.outPath} onChange={handleChange} disabled={isLoading} />
        </div>
        {/* Ffmpeg Env Var */}
        <div className="flex items-center space-x-2 pt-2">
          <Switch
            id="useFfmpegEnvVar"
            checked={config.useFfmpegEnvVar}
            onCheckedChange={handleSwitchChange}
            disabled={isLoading}
          />
          <Label htmlFor="useFfmpegEnvVar">Use FFmpeg from System Environment Variable (PATH)</Label>
        </div>
    </form>
  );
} 