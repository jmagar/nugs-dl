import React from 'react'
import { ExternalLink, HelpCircle } from "lucide-react"

const QuickHelpSection = ({ title, children }: { 
  title: string, 
  children: React.ReactNode
}) => (
  <div className="bg-gray-800/30 border border-gray-700 rounded-lg p-4">
    <h3 className="text-sm font-semibold mb-3 text-purple-300">
      {title}
    </h3>
    {children}
  </div>
)

const BulletList = ({ items }: { items: string[] }) => (
  <ul className="space-y-1 text-sm text-gray-300">
    {items.map((item, index) => (
      <li key={index} className="relative pl-4 before:content-['•'] before:absolute before:left-0 before:text-purple-400">
        {item}
      </li>
    ))}
  </ul>
)

const QuickHelp = () => {
  return (
    <div className="space-y-4">
      <div className="flex items-center gap-2">
        <HelpCircle className="h-5 w-5 text-purple-400" />
        <h2 className="text-lg font-medium">Quick Help</h2>
      </div>
      
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <QuickHelpSection title="Supported Media Types">
          <BulletList 
            items={[
              "Albums",
              "Artist catalogs", 
              "Playlists",
              "Livestreams",
              "Videos",
              "Webcasts"
            ]}
          />
        </QuickHelpSection>

        <QuickHelpSection title="URL Format Examples">
          <div className="space-y-2 text-sm text-gray-300 font-mono">
            <div>https://play.nugs.net/release/23329</div>
            <div>https://play.nugs.net/#/artist/461</div>
            <div>https://play.nugs.net/#/playlists/playlist/1234</div>
          </div>
        </QuickHelpSection>
        
        <QuickHelpSection title="Tips">
          <BulletList 
            items={[
              "Override download location for specific URLs",
              "Paste multiple URLs (one per line)", 
              "Drag and drop to prioritize downloads in queue"
            ]}
          />
          <a 
            href="https://github.com/your-repo/nugs-dl" 
            target="_blank" 
            rel="noopener noreferrer"
            className="mt-3 inline-flex items-center gap-1 text-sm text-purple-400 hover:text-purple-300 transition-colors"
          >
            View full documentation
            <ExternalLink className="h-3 w-3" />
          </a>
        </QuickHelpSection>
      </div>
    </div>
  )
}

export default QuickHelp 