<Climb>
  <header>
    <id>2Kf7</id>
    <type>feature</type>
    <description>Complete overhaul of the Nugs-Downloader Web UI to *exactly* match the provided screenshots and detailed UI/UX specifications. This includes implementing a dark theme, specific color palettes, typography, layout structure, individual component styling, responsive design, and micro-interactions.</description>
  </header>
  <newDependencies>
    - No new external libraries are anticipated beyond what's already in `webui/package.json` (React, Tailwind, shadcn/ui components). Specific icons from `lucide-react` will be used.
  </newDependencies>
  <prerequisiteChanges>
    - The existing Web UI components in `webui/src/components/` will be heavily modified or replaced.
    - Global styles in `webui/src/index.css` (or equivalent global stylesheet) will be updated.
    - Tailwind configuration (`tailwind.config.js`) might need adjustments for custom colors, fonts, and other theme elements.
  </prerequisiteChanges>
  <relevantFiles>
    - `webui/src/`: Entire directory, especially:
        - `webui/src/App.tsx` (Main application layout)
        - `webui/src/components/`: All UI components.
        - `webui/src/index.css` (Global styles)
        - `webui/tailwind.config.js` (Tailwind theme configuration)
        - `webui/src/lib/utils.ts` (Likely for `cn` utility and other helpers)
    - `webui/index.html` (For font links, base structure)
    - `webui/components.json` (shadcn/ui configuration)
  </relevantFiles>
  <everythingElse>
    ## Feature Overview: Nugs-Downloader UI Overhaul

    **Feature Name and ID:** Nugs-Downloader UI Overhaul (2Kf7)
    **Purpose Statement:** To implement a visually appealing, modern, and highly usable web interface for the Nugs-Downloader application, adhering strictly to the provided design specifications and screenshots. This will enhance user experience and provide a professional look and feel.
    **Problem Being Solved:** The current UI may not meet the desired aesthetic or functional polish. This overhaul aims to create a pixel-perfect implementation of the new design.
    **Success Metrics:**
        - The final UI matches the provided screenshots across all specified sections (Download, Queue, History, Settings).
        - All specified colors, fonts, spacing, and layout details are implemented correctly.
        - All described interactive elements (buttons, forms, tabs, modals, etc.) function as specified with correct styling and micro-interactions.
        - The UI is responsive across mobile, tablet, and desktop breakpoints as defined.
        - Accessibility considerations (keyboard navigation, ARIA labels, contrast) are implemented.

    ## Requirements

    **Functional Requirements:**
        - All existing functionality (adding downloads, managing queue, viewing history, configuring settings) must be preserved and accessible through the new UI.
        - The UI must accurately reflect the state of the backend (e.g., download progress, errors).
        - User flows for all core actions (downloading, queue management, history viewing, configuration) must match the provided descriptions.

    **Technical Requirements:**
        - Utilize the existing frontend stack: React, Vite, TypeScript, Tailwind CSS, shadcn/ui.
        - Adhere to the specified color palette, typography, and layout structure.
        - Implement all component styles as per the UI/UX description.
        - Ensure responsive design for mobile, tablet, and desktop views.
        - Implement specified micro-interactions and animations.
        - Maintain clean, readable, and maintainable code.

    **User Requirements:**
        - The interface must be intuitive and easy to navigate.
        - Visual feedback for actions and state changes must be clear.
        - The application should feel modern and responsive.

    **Constraints:**
        - Must work within the existing Go backend API structure.
        - Implementation should primarily use Tailwind CSS and shadcn/ui components, extending them as necessary.

    ## Design and Implementation

    **User Flow:** (Refer to the "User Flows" section in the provided UI/UX description document).
    **Architecture Overview:** The frontend application (React/Vite) will continue to communicate with the Go backend API. UI components will be built/modified in `webui/src/components/`. Global styles and Tailwind configuration will define the overall look and feel.
    **Dependent Components:** Primarily relies on `lucide-react` for icons and `@radix-ui/*` components via shadcn/ui.
    **API Specifications:** No changes to the backend API are anticipated for this UI-only overhaul.
    **Data Models:** Frontend data models for representing downloads, queue items, history, and configuration will remain largely the same, but their presentation will change.

    ## Development Details

    **Relevant Files:** (Listed in `<relevantFiles>` tag above).
    **Implementation Considerations:**
        - Start with global styles: background, typography, color variables in Tailwind config.
        - Implement the main layout: Header, Footer, main content area.
        - Build/style shared components first (Buttons, Inputs, Cards, etc.).
        - Implement each main tab/section (Download Manager, Queue Manager, History) one by one.
        - Implement the Configuration Modal.
        - Address responsiveness and micro-interactions throughout the process.
    **Dependencies:** (Listed in `<newDependencies>` tag above).
    **Security Considerations:** Not directly impacted by UI changes, but ensure no new vulnerabilities are introduced (e.g., XSS if dynamic content is handled improperly, though unlikely with React).

    ## Testing Approach

    **Test Cases:**
        - Visual comparison against screenshots for all sections and components.
        - Functional testing of all user flows.
        - Responsiveness testing on different screen sizes.
        - Interaction testing for all buttons, forms, and dynamic elements.
        - Accessibility checks (keyboard navigation, screen reader compatibility).
    **Acceptance Criteria:**
        - UI is pixel-perfect according to screenshots and specifications.
        - All functionalities are working as expected.
        - No visual regressions or bugs.
    **Edge Cases:**
        - Empty states for Queue and History.
        - Very long text in titles or metadata fields.
        - Error states and their display.
    **Performance Requirements:**
        - UI should load quickly and feel responsive.
        - Animations should be smooth (60fps).

    ## Design Assets

    **Mockups/Wireframes:** The 7 provided screenshots serve as the primary design assets.
    **User Interface Guidelines:** The detailed "Comprehensive UI/UX Description" document.
    **Content Guidelines:** Use placeholder text from screenshots or sensible defaults where applicable.

    ## Future Considerations

    **Scalability Plans:** The component-based architecture should allow for easier future modifications.
    **Enhancement Ideas:** (Outside current scope) Could include more advanced filtering/sorting, bulk actions, or user customization.
    **Known Limitations:** Dependent on the stability and correctness of the backend API.

  </everythingElse>
</Climb> 