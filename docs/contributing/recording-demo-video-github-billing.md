# Recording Demo Video for GitHub Billing Usage Component

This guide explains how to record a demo video showing the GitHub Billing Usage component in action. This video is required as part of the PR submission for integration bounties.

## Video Requirements

Your video should demonstrate:
1. Setting up the GitHub integration with the new Administration permission
2. Configuring and running the Get Billing Usage component
3. Viewing the output with billing usage data

## Before You Start

### Prerequisites

- A GitHub organization (not a personal account) with:
  - At least one private repository
  - GitHub Actions enabled and some usage history
  - Organization owner/admin access to approve new permissions
- A working SuperPlane instance (local development or staging)
- Screen recording software (see recommendations below)

### Screen Recording Tools

Choose one of these tools based on your operating system:

**macOS:**
- QuickTime Player (built-in, free)
- ScreenFlow (paid, professional)
- OBS Studio (free, open source)

**Windows:**
- Xbox Game Bar (built-in, free)
- OBS Studio (free, open source)
- Camtasia (paid, professional)

**Linux:**
- SimpleScreenRecorder (free)
- OBS Studio (free, open source)
- Kazam (free)

## Recording Steps

### Part 1: Setting up the Integration (~2-3 minutes)

1. **Start Recording** - Begin your screen recording

2. **Navigate to Integrations**
   - Open your SuperPlane instance
   - Go to Settings → Integrations
   - Click "Add Integration"

3. **Configure GitHub Integration**
   - Select "GitHub" from the integration list
   - Enter your organization name (if required)
   - Click "Continue" or equivalent button

4. **Create GitHub App**
   - The system will redirect you to GitHub
   - Give your app a name (e.g., "SuperPlane Test")
   - Click "Create GitHub App"
   - **Important**: Show the permissions screen where "Administration: Read" is listed

5. **Install GitHub App**
   - Select your organization
   - Choose which repositories to grant access (select at least one private repository)
   - Click "Install"

6. **Approve New Permission** (if updating existing integration)
   - If you already have a SuperPlane GitHub app, you'll see a permission request
   - Show the "Administration: Read" permission being requested
   - Click "Accept new permissions"

7. **Verify Integration is Ready**
   - Return to SuperPlane
   - Show that the integration status is "Ready" or "Connected"
   - Optionally show the list of accessible repositories

### Part 2: Using the Get Billing Usage Component (~3-4 minutes)

1. **Create a New Workflow**
   - Navigate to Workflows
   - Click "Create New Workflow"
   - Give it a descriptive name (e.g., "GitHub Billing Usage Demo")

2. **Add Get Billing Usage Component**
   - Open the component palette
   - Search for "GitHub" or navigate to GitHub integration components
   - Drag "Get Billing Usage" onto the canvas

3. **Configure the Component**
   - Click on the Get Billing Usage component to open configuration
   - **Show each configuration option:**
     - **Repositories**: Leave empty for org-wide, or select specific repositories
     - **Year**: Set to current year (or leave default)
     - **Month**: Set to current month (or leave default)
     - **Day**: Leave empty for monthly summary
     - **Product**: Leave as "actions" (default)
     - **Runner OS / SKU**: Leave empty or select specific OS (e.g., UBUNTU)
   
4. **Add Debug/Log Component** (optional but recommended)
   - Connect the output of Get Billing Usage to a Debug or Log component
   - This will make it easier to see the output in the demo

5. **Run the Workflow**
   - Click "Run" or "Test"
   - Show the workflow executing

6. **Display the Results**
   - Show the execution log or output panel
   - **Highlight the key output fields:**
     - `minutes_used`: Total billable minutes
     - `minutes_used_breakdown`: Breakdown by OS (Linux, Windows, macOS)
     - `total_cost`: Cost if available
     - `repository_breakdown`: Per-repository usage if multiple repos selected
   
7. **Try Different Configurations** (optional, ~2 minutes)
   - Edit the component to filter by specific repositories
   - Run again and show the filtered results
   - Edit to filter by specific OS/SKU
   - Run again and show the filtered results

### Part 3: Wrap Up (~1 minute)

1. **Show Final Output**
   - Display a clear view of the complete output data
   - Pause for a few seconds so viewers can read the data

2. **Summary** (optional voice-over)
   - Briefly mention what was demonstrated
   - Note any special considerations (e.g., "Only private repos show billable usage")

3. **Stop Recording**

## Tips for a Great Video

### Video Quality

- **Resolution**: 1920x1080 (1080p) minimum
- **Frame Rate**: 30 fps minimum
- **Audio**: Optional, but helpful narration can enhance the video
- **Length**: 5-10 minutes total is ideal

### Recording Best Practices

1. **Clean Your Desktop**: Close unnecessary applications and windows
2. **Zoom In**: Use zoom or large fonts so text is readable
3. **Go Slowly**: Move your mouse deliberately and pause after clicks
4. **Show, Don't Tell**: Visual actions are more important than narration
5. **Highlight Important Elements**: Use your mouse to point at key UI elements
6. **Test First**: Do a dry run before the actual recording

### Common Mistakes to Avoid

- ❌ Recording too quickly - viewers need time to see what you're doing
- ❌ Small fonts or resolution - text should be easily readable
- ❌ Skipping the permission approval step - this is a key part of the feature
- ❌ Not showing the actual output data - the results are important
- ❌ Video longer than 15 minutes - keep it concise

## Editing Your Video

### Recommended Edits

1. **Trim Dead Time**: Remove long pauses or loading screens
2. **Add Annotations**: Consider adding text overlays to highlight important steps
3. **Speed Up Slow Parts**: Use 2x speed for slow loading or navigation
4. **Add Intro/Outro**: Brief title card and ending screen (optional)

### Editing Tools

- **macOS**: iMovie (free), Final Cut Pro (paid)
- **Windows**: Windows Video Editor (free), Adobe Premiere (paid)
- **Linux**: Kdenlive (free), DaVinci Resolve (free)
- **Online**: Kapwing, Clipchamp

## Uploading Your Video

1. **Export Settings**
   - Format: MP4 (H.264 codec)
   - Quality: High/Best
   - Keep the file size under 100MB if possible

2. **Upload Platforms**
   - **YouTube**: Upload as unlisted and share the link
   - **Loom**: Quick screen recording and sharing
   - **Google Drive**: Share with link permissions
   - **Vimeo**: Professional hosting option

3. **Share the Link**
   - Include the video link in your PR description
   - Make sure the link is publicly accessible
   - Test the link in an incognito/private browser window

## Example Video Structure

```
[00:00-00:30] Introduction
- Show SuperPlane dashboard
- Navigate to Integrations

[00:30-02:30] Setting Up Integration
- Configure GitHub integration
- Create GitHub App
- Install app in organization
- Show Administration permission
- Verify connection

[02:30-05:00] Using Get Billing Usage Component
- Create workflow
- Add component
- Configure settings
- Run workflow

[05:00-06:30] Viewing Results
- Show output data
- Highlight key metrics
- Try different configurations

[06:30-07:00] Conclusion
- Show final results
- End recording
```

## Troubleshooting

### No Billing Data Showing

- **Cause**: Only private repositories on GitHub-hosted runners show billable usage
- **Solution**: Ensure you're testing with a private repository that has run GitHub Actions

### Permission Denied (403 Error)

- **Cause**: Administration permission not yet approved
- **Solution**: Go to GitHub → Settings → Installed GitHub Apps → [Your App] → Permissions and approve the new permission

### No Usage Data for Current Month

- **Cause**: No Actions have run yet this month
- **Solution**: Either run some Actions or query a previous month with usage

## Need Help?

If you encounter issues while recording your video:

1. Join our [Discord](https://discord.gg/KC78eCNsnw) and ask in the #support channel
2. Check the [Contributing Guide](../../CONTRIBUTING.md) for general setup help
3. Review the [Integration PRs Guide](integration-prs.md) for PR requirements

## Video Submission Checklist

Before submitting your PR, verify:

- [ ] Video shows GitHub App installation with Administration permission
- [ ] Video demonstrates Get Billing Usage component configuration
- [ ] Video shows actual billing data output
- [ ] Video is 5-10 minutes long
- [ ] Video resolution is at least 1080p
- [ ] Video link is publicly accessible
- [ ] Video link is included in PR description
