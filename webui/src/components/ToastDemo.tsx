// import React from 'react'
import { Button } from "@/components/ui/button"
import { toast } from 'sonner'
import { 
  Bell, 
  CheckCircle, 
  XCircle, 
  AlertCircle, 
  Info,
  Loader2
} from "lucide-react"
import { FOCUS_STYLES, createStatusAnnouncement } from '@/lib/accessibility'
import { INTERACTIVE } from '@/lib/animations'

const ToastDemo = () => {
  // Helper function to show a normal toast
  const showDefaultToast = () => {
    toast('Default notification', {
      description: 'This is a simple toast notification'
    });
    createStatusAnnouncement('Showing default toast notification');
  }

  // Helper function to show a success toast
  const showSuccessToast = () => {
    toast.success('Success!', {
      description: 'Operation completed successfully'
    });
    createStatusAnnouncement('Showing success toast notification');
  }

  // Helper function to show an error toast
  const showErrorToast = () => {
    toast.error('Error!', {
      description: 'Something went wrong. Please try again.'
    });
    createStatusAnnouncement('Showing error toast notification');
  }

  // Helper function to show a warning toast
  const showWarningToast = () => {
    toast.warning('Warning!', {
      description: 'This action might have consequences'
    });
    createStatusAnnouncement('Showing warning toast notification');
  }

  // Helper function to show an info toast
  const showInfoToast = () => {
    toast.info('Information', {
      description: 'Here is some useful information for you'
    });
    createStatusAnnouncement('Showing info toast notification');
  }

  // Helper function to show a toast with action
  const showActionToast = () => {
    toast('Action required', {
      description: 'Please confirm or cancel this action',
      action: {
        label: 'Confirm',
        onClick: () => {
          toast.success('Action confirmed!');
          createStatusAnnouncement('Action confirmed');
        }
      },
      cancel: {
        label: 'Cancel',
        onClick: () => {
          toast.error('Action cancelled');
          createStatusAnnouncement('Action cancelled');
        }
      }
    });
    createStatusAnnouncement('Showing action toast notification with confirm and cancel buttons');
  }

  // Helper function to show a promise toast
  const showPromiseToast = () => {
    const promise = new Promise((resolve, reject) => {
      // Simulate API call
      setTimeout(() => {
        if (Math.random() > 0.3) {
          resolve('Data successfully loaded');
          createStatusAnnouncement('Data loaded successfully');
        } else {
          reject(new Error('Failed to load data'));
          createStatusAnnouncement('Failed to load data', 'assertive');
        }
      }, 2000);
    });

    toast.promise(promise, {
      loading: 'Loading data...',
      success: (data) => `${data}`,
      error: (err) => `Error: ${err.message}`
    });
    
    createStatusAnnouncement('Loading data, please wait...');
  }

  return (
    <div className="flex flex-col gap-4 p-4 rounded-lg border border-border-primary" role="region" aria-labelledby="toast-demo-title">
      <h2 id="toast-demo-title" className="text-lg font-semibold mb-2">Toast Notifications</h2>
      <div className="grid grid-cols-2 md:grid-cols-3 gap-3">
        <Button
          variant="outline"
          className={`flex items-center justify-center gap-2 ${INTERACTIVE.button} ${FOCUS_STYLES.BUTTON}`}
          onClick={showDefaultToast}
          aria-label="Show default toast notification"
        >
          <Bell className="h-4 w-4" aria-hidden="true" />
          <span>Default</span>
        </Button>
        
        <Button
          variant="outline"
          className={`flex items-center justify-center gap-2 ${INTERACTIVE.button} ${FOCUS_STYLES.BUTTON}`}
          onClick={showSuccessToast}
          aria-label="Show success toast notification"
        >
          <CheckCircle className="h-4 w-4 text-success" aria-hidden="true" />
          <span>Success</span>
        </Button>
        
        <Button
          variant="outline"
          className={`flex items-center justify-center gap-2 ${INTERACTIVE.button} ${FOCUS_STYLES.BUTTON}`}
          onClick={showErrorToast}
          aria-label="Show error toast notification"
        >
          <XCircle className="h-4 w-4 text-error" aria-hidden="true" />
          <span>Error</span>
        </Button>
        
        <Button
          variant="outline"
          className={`flex items-center justify-center gap-2 ${INTERACTIVE.button} ${FOCUS_STYLES.BUTTON}`}
          onClick={showWarningToast}
          aria-label="Show warning toast notification"
        >
          <AlertCircle className="h-4 w-4 text-warning" aria-hidden="true" />
          <span>Warning</span>
        </Button>
        
        <Button
          variant="outline"
          className={`flex items-center justify-center gap-2 ${INTERACTIVE.button} ${FOCUS_STYLES.BUTTON}`}
          onClick={showInfoToast}
          aria-label="Show info toast notification"
        >
          <Info className="h-4 w-4 text-info" aria-hidden="true" />
          <span>Info</span>
        </Button>
        
        <Button
          variant="outline"
          className={`flex items-center justify-center gap-2 ${INTERACTIVE.button} ${FOCUS_STYLES.BUTTON}`}
          onClick={showActionToast}
          aria-label="Show toast with action buttons"
        >
          <span>Action</span>
        </Button>
        
        <Button
          variant="outline"
          className={`flex items-center justify-center gap-2 col-span-full md:col-span-1 ${INTERACTIVE.button} ${FOCUS_STYLES.BUTTON}`}
          onClick={showPromiseToast}
          aria-label="Show promise-based loading toast notification"
        >
          <Loader2 className="h-4 w-4 mr-1 animate-spin" aria-hidden="true" />
          <span>Promise</span>
        </Button>
      </div>
    </div>
  )
}

export default ToastDemo 