# Skills Hub GitHub Importer & AI Dependency Analysis Architecture Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:**
1. Remove obsolete **Test Bridge** button from the top navigation bar in `App.svelte`.
2. Expand **Skills Hub** (`SkillCatalog.svelte`) with a dedicated **"Add Skills from GitHub"** flow:
   - Accept any GitHub repository URL (e.g. `https://github.com/owner/repo` or `owner/repo`).
   - Clone or scan repository trees for SKILL.md definitions, commands, and recipes.
   - Display a searchable checklist of discovered skills with "Select All", per-skill checkboxes, categories, and tags.
   - Run an AI requirement & prerequisite analysis on selected skills (evaluating README.md, package.json / requirements.txt, and CLI dependencies like `node`, `python`, `go`, or system tools).
   - Provide an interactive terminal execution / install log viewer with confirmation before running setup commands on the local machine.

**Architecture & Components:**
- **Top Bar**: Remove `<Button>Test Bridge</Button>` from `frontend/src/App.svelte` and clean up `testBridge()` helper.
- **Backend (Go)**: In `pkg/grokrunner` and `app.go`:
  - `ScanGitHubRepoSkills(repoUrl string)`: Clones to `/tmp/grok-skills-temp` (or fetches tree via GitHub API/git clone) and parses all `SKILL.md` / `*.skill.md` files.
  - `AnalyzeSkillPrerequisites(repoPath string, skillPaths []string)`: Analyzes dependencies (Node packages, Python modules, binary tools) needed by the selected skills.
  - `InstallDiscoveredSkills(repoPath string, skillPaths []string, targetScope string)`: Copies selected skill folders into `~/.grok/skills/` or `~/.agents/skills/`.
  - `ExecuteSkillSetupCommands(commands []string)`: Streams terminal execution logs for npm/pip/brew install commands with real-time output events.
- **Frontend (Svelte 5 Runes & Ant Design Dark Tokens)**:
  - Add "Add Skill (GitHub)" button in `SkillCatalog.svelte`.
  - Step 1: Input GitHub repo URL with auto-clean and validation.
  - Step 2: Skill selection table with checkboxes, select-all, badges, description preview, and requirement chips.
  - Step 3: AI Prerequisite & Installation Plan view (showing recommended shell commands).
  - Step 4: Live installation log output card with success confirmation and refresh into local skill library.

---

### Task 1: Clean Up Header Navigation (Remove Test Bridge)

**Files:**
- Modify: `frontend/src/App.svelte`

- [ ] **Step 1: Remove Test Bridge button and helper**
Remove Test Bridge `<Button>` from the header and delete `testBridge()` method.

---

### Task 2: Backend GitHub Repo Scanner & Skill Installer in Go

**Files:**
- Modify: `pkg/grokrunner/skill_installer.go` (create new)
- Modify: `app.go`
- Modify: `frontend/src/app.d.ts`

**Data Models:**
```go
type DiscoveredSkill struct {
    Name         string   `json:"name"`
    Description  string   `json:"description"`
    Category     string   `json:"category"`
    RelativePath string   `json:"relativePath"`
    Prereqs      []string `json:"prereqs"`
    Commands     []string `json:"commands"`
}

type SkillAnalysisResult struct {
    RepoName         string            `json:"repoName"`
    TempPath         string            `json:"tempPath"`
    Skills           []DiscoveredSkill `json:"skills"`
    GlobalPrereqs    []string          `json:"globalPrereqs"`
    SuggestedScripts []string          `json:"suggestedScripts"`
}
```

- [ ] **Step 1: Implement `ScanGitHubRepoSkills` in `pkg/grokrunner/skill_installer.go`**
Clone the git repository to a temporary directory under `/tmp/aethergrok-skills/`, scan for all `SKILL.md` files, parse frontmatter metadata (`name`, `description`, `category`), inspect README/package.json for required runtimes/tools.
- [ ] **Step 2: Implement `InstallDiscoveredSkills` and `ExecuteSkillSetupCommands`**
Copy selected skills into `~/.grok/skills/<skill_name>/` and stream command execution output.
- [ ] **Step 3: Expose Go methods in `app.go` and update TypeScript declarations in `frontend/src/app.d.ts`**

---

### Task 3: Interactive Skills Hub Importer UI in `SkillCatalog.svelte`

**Files:**
- Modify: `frontend/src/lib/components/skills/SkillCatalog.svelte`
- Create: `frontend/src/lib/components/skills/SkillImporterModal.svelte`

**UI/UX Design Flow:**
1. **Header Action**: "+ Import from GitHub" button beside Search bar.
2. **Step 1: Repository Input Modal**:
   - Dark theme input field for GitHub URL (supports `owner/repo` or full HTTPS URL).
   - "Analyze Repository" button with spinning AI loader.
3. **Step 2: Skill Multi-Select Grid**:
   - Checkbox header (`Select All (N skills)`).
   - List of discovered skills with tags, descriptions, and file locations.
   - AI Analysis Card displaying detected system dependencies (e.g. `node >= 18`, `python3`, `ripgrep`).
4. **Step 3: Execution & Installation Terminal Drawer**:
   - Monospace live log viewer showing copying of skill folders and execution of setup scripts.
   - Confirmation dialog before running local shell setup scripts.
   - Successful completion notification with automatic refresh of Skills Hub catalog.

---

### Task 4: Verification & Build

- [ ] **Step 1: Verify TypeScript and Vite builds**
- [ ] **Step 2: Run Wails clean build and test in desktop application**
- [ ] **Step 3: Test importing a real GitHub skills repository**
