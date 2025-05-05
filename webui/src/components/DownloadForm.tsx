import React, { useState } from 'react';
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Checkbox } from "@/components/ui/checkbox"; // Using Checkbox for flags
import { toast } from "sonner";

// Type matching api.DownloadOptions for clarity
type DownloadOptionsState = {
  forceVideo: boolean;
  skipVideos: boolean;
  skipChapters: boolean;
};

export function DownloadForm() {
  const [urls, setUrls] = useState<string>("");
  const [options, setOptions] = useState<DownloadOptionsState>({
    forceVideo: false,
    skipVideos: false,
    skipChapters: false,
  });
  const [isSubmitting, setIsSubmitting] = useState(false);

  const handleUrlChange = (event: React.ChangeEvent<HTMLTextAreaElement>) => {
    setUrls(event.target.value);
  };

  const handleOptionChange = (optionKey: keyof DownloadOptionsState) => {
    setOptions(prevOptions => ({
      ...prevOptions,
      [optionKey]: !prevOptions[optionKey],
    }));
  };

  const handleSubmit = async (event: React.FormEvent) => {
    event.preventDefault();
    setIsSubmitting(true);
    
    const urlList = urls.split(/\r?\n/).map(url => url.trim()).filter(url => url !== '');
    
    if (urlList.length === 0) {
      toast.error("Please enter at least one URL.");
      setIsSubmitting(false);
      return;
    }

    const payload = {
      urls: urlList,
      // Ensure options match the structure expected by pkg/api/types.go AddDownloadRequest
      options: {
          forceVideo: options.forceVideo,
          skipVideos: options.skipVideos,
          skipChapters: options.skipChapters,
      },
    };

    console.log("Submitting download request:", payload);
    toast.info("Adding job(s) to queue...");

    try {
        const response = await fetch('/api/downloads', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(payload),
        });

        // Check if the request was accepted (202) or failed
        if (!response.ok) {
            let errorMsg = `HTTP error! status: ${response.status}`;
            try {
                 // Try to get more specific error from backend response
                const errorData = await response.json();
                errorMsg = errorData.error || errorMsg;
            } catch (jsonError) {
                // Backend didn't send JSON error, use status text
                errorMsg = response.statusText || errorMsg;
            }
            throw new Error(errorMsg);
        }

        // Request was accepted (202)
        const jobData = await response.json(); // Get the job details returned by the API
        toast.success(`Job added to queue successfully! ID: ${jobData.id}`);
        setUrls(""); // Clear input on success
        // Optionally reset options:
        // setOptions({ forceVideo: false, skipVideos: false, skipChapters: false });

    } catch (err: unknown) {
        console.error("Error submitting download request:", err);
        const errorMessage = err instanceof Error ? err.message : 'Unknown error occurred';
        toast.error(`Failed to add job: ${errorMessage}`);
    } finally {
        setIsSubmitting(false);
    }
  };

  return (
    <Card className="w-full mt-6"> {/* Add margin top */} 
      <CardHeader>
        <CardTitle>Add Downloads</CardTitle>
        <CardDescription>Enter nugs.net URLs (one per line) to add them to the download queue.</CardDescription>
      </CardHeader>
      <form onSubmit={handleSubmit}>
        <CardContent className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="urls">URLs / File Paths</Label>
            <Textarea
              id="urls"
              placeholder="https://play.nugs.net/release/...
https://play.nugs.net/#/videos/artist/..."
              rows={5}
              value={urls}
              onChange={handleUrlChange}
              disabled={isSubmitting}
            />
          </div>
          <div className="space-y-2 pt-2">
             <Label>Download Options</Label>
             <div className="flex items-center space-x-2">
               <Checkbox 
                 id="forceVideo"
                 checked={options.forceVideo}
                 onCheckedChange={() => handleOptionChange('forceVideo')}
                 disabled={isSubmitting}
               />
               <Label htmlFor="forceVideo" className="text-sm font-normal cursor-pointer">Force Video (if available)</Label>
             </div>
             <div className="flex items-center space-x-2">
               <Checkbox 
                 id="skipVideos"
                 checked={options.skipVideos}
                 onCheckedChange={() => handleOptionChange('skipVideos')}
                 disabled={isSubmitting}
                />
               <Label htmlFor="skipVideos" className="text-sm font-normal cursor-pointer">Skip Videos (in artist URLs)</Label>
             </div>
             <div className="flex items-center space-x-2">
               <Checkbox 
                 id="skipChapters"
                 checked={options.skipChapters}
                 onCheckedChange={() => handleOptionChange('skipChapters')}
                 disabled={isSubmitting}
                />
               <Label htmlFor="skipChapters" className="text-sm font-normal cursor-pointer">Skip Video Chapters</Label>
             </div>
          </div>
        </CardContent>
        <CardFooter>
          <Button type="submit" disabled={isSubmitting || urls.trim() === ''}>
            {isSubmitting ? 'Adding to Queue...' : 'Add to Download Queue'}
          </Button>
        </CardFooter>
      </form>
    </Card>
  );
} 