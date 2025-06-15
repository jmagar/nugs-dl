// import React from 'react'; // React and useState removed as they are no longer used
import { Button } from "@/components/ui/button";
// Other imports (Label, Textarea, etc.) are still here but unused, which is fine for the test.
// import { Label } from "@/components/ui/label";
// import { Textarea } from "@/components/ui/textarea";
// import { Checkbox } from "@/components/ui/checkbox";
// import { Video, VideoOff, Scissors } from 'lucide-react';
// import { toast } from "sonner";

// Type matching api.DownloadOptions for clarity
// type DownloadOptionsState = {
//   forceVideo: boolean;
//   skipVideos: boolean;
//   skipChapters: boolean;
// };

export function DownloadForm() {
  // State and handlers are kept for now, but not used in this super-simple return
  // const [urls, setUrls] = useState<string>("");
  // const [options, setOptions] = useState<DownloadOptionsState>({
//   forceVideo: false,
//   skipVideos: false,
//   skipChapters: false,
// });
  // const [isSubmitting, setIsSubmitting] = useState(false);

  // const handleUrlChange = (event: React.ChangeEvent<HTMLTextAreaElement>) => {
  //   setUrls(event.target.value);
  // };

  // const handleOptionChange = (optionKey: keyof DownloadOptionsState) => {
  //   setOptions(prevOptions => ({
  //     ...prevOptions,
  //     [optionKey]: !prevOptions[optionKey],
  //   }));
  // };

  // const handleSubmit = async (event: React.FormEvent) => {
  //   event.preventDefault();
  //   setIsSubmitting(true);
  //   const urlList = urls.split(/\r?\n/).map(url => url.trim()).filter(url => url !== '');
  //   if (urlList.length === 0) {
  //     toast.error("Please enter at least one URL.");
  //   } finally {
  //       setIsSubmitting(false);
  //   }
  // };

  // Test with Shadcn Button only
  return (
    <Button 
      type="submit"
      className="w-full bg-accent-blue text-white hover:bg-accent-blue/90 mt-4"
      disabled={true} // Temporarily disabled as isSubmitting and urls are commented out
    >
      'Add to Download Queue' // Temporarily set as isSubmitting is commented out
    </Button>
  );
} 