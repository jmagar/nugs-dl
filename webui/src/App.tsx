import React, { useState, useEffect, useCallback } from 'react';
import { Toaster } from "@/components/ui/sonner"
import { ConfigForm } from "@/components/ConfigForm"
import { DownloadForm } from "@/components/DownloadForm"
import { QueueDisplay } from "@/components/QueueDisplay"
import { Button } from "@/components/ui/button"
import { Sheet, SheetContent, SheetDescription, SheetHeader, SheetTitle, SheetTrigger, SheetFooter, SheetClose } from "@/components/ui/sheet"
import { Settings, Moon, Sun } from "lucide-react"
import { useTheme } from "next-themes"
import { toast } from "sonner"
import { type AppConfig } from "@/types/api"

function ThemeToggle() {
  const { setTheme, theme } = useTheme();

  return (
    <Button
      variant="ghost"
      size="icon"
      onClick={() => setTheme(theme === "light" ? "dark" : "light")}
      aria-label="Toggle theme"
    >
      <Sun className="h-[1.2rem] w-[1.2rem] rotate-0 scale-100 transition-all dark:-rotate-90 dark:scale-0" />
      <Moon className="absolute h-[1.2rem] w-[1.2rem] rotate-90 scale-0 transition-all dark:rotate-0 dark:scale-100" />
    </Button>
  );
}

function App() {
  const [appConfig, setAppConfig] = useState<AppConfig | null>(null);
  const [isConfigLoading, setIsConfigLoading] = useState(true);
  const [configError, setConfigError] = useState<string | null>(null);
  const [isSavingConfig, setIsSavingConfig] = useState(false);

  useEffect(() => {
    let isMounted = true;
    const fetchConfig = async () => {
      setIsConfigLoading(true);
      setConfigError(null);
      try {
        const response = await fetch('/api/config');
        if (!response.ok) {
          throw new Error(`Failed to fetch config: ${response.statusText}`);
        }
        const data: AppConfig = await response.json();
        if (isMounted) setAppConfig(data);
      } catch (err: unknown) {
        if (isMounted) {
            const msg = err instanceof Error ? err.message : 'Failed to load configuration.';
            setConfigError(msg);
            toast.error(msg);
        }
      } finally {
        if (isMounted) setIsConfigLoading(false);
      }
    };
    fetchConfig();
    return () => { isMounted = false; }
  }, []);

  const handleSaveConfig = useCallback(async (configToSave: AppConfig) => {
    setIsSavingConfig(true);
    setConfigError(null);
    toast.info("Saving configuration...");
    try {
      const response = await fetch('/api/config', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(configToSave),
      });
      if (!response.ok) {
         let errorMsg = `HTTP error! status: ${response.status}`;
         try { 
             const errorData = await response.json(); 
             errorMsg = errorData.error || errorMsg; 
         } catch (jsonErr) { 
            // Ignore JSON parsing error, keep the status-based message
            console.warn("Could not parse error response body:", jsonErr);
         } 
         throw new Error(errorMsg);
      }
      const savedConfig: AppConfig = await response.json();
      setAppConfig(savedConfig);
      toast.success("Configuration saved successfully!");
    } catch (err: unknown) {
        const msg = err instanceof Error ? err.message : 'Unknown error occurred';
        setConfigError(msg);
        toast.error(`Failed to save configuration: ${msg}`);
    } finally {
        setIsSavingConfig(false);
    }
  }, []);

  return (
    <div className="min-h-screen bg-background text-foreground flex flex-col items-center p-4 md:p-8">
      <header className="w-full max-w-4xl mb-6 flex justify-between items-center">
        <h1 className="text-2xl md:text-3xl font-bold">Nugs DL Web UI</h1>
        <div className="flex items-center gap-2">
          <Sheet>
            <SheetTrigger asChild>
              <Button variant="ghost" size="icon" aria-label="Open Settings">
                <Settings className="h-[1.2rem] w-[1.2rem]" />
              </Button>
            </SheetTrigger>
            <SheetContent>
              <SheetHeader>
                <SheetTitle>Configuration</SheetTitle>
                <SheetDescription>
                  Manage nugs.net credentials and default download settings.
                  {configError && <p className="text-sm font-medium text-destructive pt-2">Error: {configError}</p>}
                </SheetDescription>
              </SheetHeader>
              <div className="py-4">
                <ConfigForm 
                  key={JSON.stringify(appConfig)}
                  initialConfig={appConfig} 
                  isLoading={isConfigLoading || isSavingConfig}
                  onSubmit={handleSaveConfig}
                />
              </div>
              <SheetFooter>
                <SheetClose asChild>
                    <Button variant="outline">Cancel</Button>
                </SheetClose>
                <Button type="submit" form="config-form" disabled={isSavingConfig}>
                    {isSavingConfig ? "Saving..." : "Save Changes"}
                </Button>
              </SheetFooter>
            </SheetContent>
          </Sheet>
          <ThemeToggle />
        </div>
      </header>
      <main className="w-full max-w-4xl flex-grow space-y-8">
        <DownloadForm />
        <QueueDisplay />
      </main>
      <footer className="w-full max-w-4xl mt-12 text-center text-xs text-muted-foreground">
        {/* Footer content if needed */}
      </footer>
      <Toaster richColors />
    </div>
  )
}

export default App
