import React, { useState } from 'react';
import { Button } from "@/components/ui/button";
// Other imports (Label, Textarea, etc.) are still here but unused, which is fine for the test.
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Checkbox } from "@/components/ui/checkbox";
import { Video, VideoOff, Scissors } from 'lucide-react';
import { toast } from "sonner";

// Type matching api.DownloadOptions for clarity
type DownloadOptionsState = {
  forceVideo: boolean;
  skipVideos: boolean;
  skipChapters: boolean;
};

export function DownloadForm() {
  // State and handlers are kept for now, but not used in this super-simple return
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
        if (!response.ok) {
            let errorMsg = `HTTP error! status: ${response.status}`;
            try {
                const errorData = await response.json();
                errorMsg = errorData.error || errorMsg;
            } catch {
                errorMsg = response.statusText || errorMsg;
            }
            throw new Error(errorMsg);
        }
        const jobData = await response.json();
        toast.success(`Job added to queue successfully! ID: ${jobData.id}`);
        setUrls("");
    } catch (err: unknown) {
        console.error("Error submitting download request:", err);
        const errorMessage = err instanceof Error ? err.message : 'Unknown error occurred';
        toast.error(`Failed to add job: ${errorMessage}`);
    } finally {
        setIsSubmitting(false);
    }
  };

  // Test with Shadcn Button only
  return (
    <Button 
      type="submit"
      className="w-full bg-accent-blue text-white hover:bg-accent-blue/90 mt-4"
      disabled={isSubmitting || urls.trim() === ''}
    >
      {isSubmitting ? 'Adding to Queue...' : 'Add to Download Queue'}
    </Button>
  );
} 