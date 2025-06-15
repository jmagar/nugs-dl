// import React from 'react'
import { Badge } from '@/components/ui/badge'
import { 
  PlayCircle, 
  Music, 
  Calendar, 
  ListMusic,
  Download,
  Clock,
  Pause,
  CheckCircle,
  AlertCircle
} from 'lucide-react'

const BadgeExamples = () => {
  return (
    <div className="space-y-6 p-4">
      <div>
        <h3 className="text-h3 font-h3 text-text-primary mb-2">Content Type Badges</h3>
        <div className="flex flex-wrap gap-2">
          <Badge variant="album"><Music /> Album</Badge>
          <Badge variant="video"><PlayCircle /> Video</Badge>
          <Badge variant="livestream"><Calendar /> Livestream</Badge>
          <Badge variant="playlist"><ListMusic /> Playlist</Badge>
        </div>
      </div>
      
      <div>
        <h3 className="text-h3 font-h3 text-text-primary mb-2">Status Badges</h3>
        <div className="flex flex-wrap gap-2">
          <Badge variant="downloading"><Download /> Downloading</Badge>
          <Badge variant="queued"><Clock /> Queued</Badge>
          <Badge variant="paused"><Pause /> Paused</Badge>
          <Badge variant="completed"><CheckCircle /> Completed</Badge>
          <Badge variant="error"><AlertCircle /> Error</Badge>
        </div>
      </div>
      
      <div>
        <h3 className="text-h3 font-h3 text-text-primary mb-2">Legacy Badges</h3>
        <div className="flex flex-wrap gap-2">
          <Badge variant="default">Default</Badge>
          <Badge variant="secondary">Secondary</Badge>
          <Badge variant="destructive">Destructive</Badge>
          <Badge variant="outline">Outline</Badge>
        </div>
      </div>
    </div>
  )
}

export default BadgeExamples 