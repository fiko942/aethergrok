<script lang="ts">
  import { tick, onMount, onDestroy } from 'svelte';
  import type { SkillAnalysisResult, DiscoveredSkill, SkillInstallPayload, SkillInstallResult } from '../../../app.d';
  import Button from '$lib/antd/Button.svelte';
  import {
    Sparkles,
    GitBranch,
    FolderGit2,
    Check,
    Square,
    CheckSquare,
    AlertCircle,
    Loader2,
    Terminal,
    ChevronRight,
    Download,
    X,
    ExternalLink,
    Code,
    RefreshCw,
    Layers,
    CheckCircle2,
    Wrench
  } from 'lucide-svelte';

  interface Props {
    visible: boolean;
    onClose: () => void;
    onInstalled: () => void;
  }

  let { visible = $bindable(false), onClose, onInstalled }: Props = $props();

  // Wizard Steps: 1: Repo URL Input, 2: Skill Selection & Prereq Inspection, 3: Install & Terminal Logs
  let currentStep = $state<1 | 2 | 3>(1);

  // Step 1: Input State
  let repoUrl = $state('');
  let isScanning = $state(false);
  let scanError = $state<string | null>(null);

  // Step 2: Discovered Skills & Selection State
  let analysisResult = $state<SkillAnalysisResult | null>(null);
  let selectedSkillPaths = $state<string[]>([]);
  let targetScope = $state<'grok' | 'agents'>('grok');
  let runSuggestedScripts = $state(true);

  // Step 3: Installation & Real-Time Terminal Execution State
  let isInstalling = $state(false);
  let installLogs = $state<string[]>([]);
  let installResult = $state<SkillInstallResult | null>(null);
  let terminalContainerRef = $state<HTMLDivElement | null>(null);
  let unsubLog: (() => void) | undefined;

  onMount(() => {
    if (window.runtime?.EventsOn) {
      unsubLog = window.runtime.EventsOn('skill:install_log', (line: string) => {
        installLogs = [...installLogs, line];
        tick().then(scrollTerminal);
      });
    }
  });

  onDestroy(() => {
    if (unsubLog) {
      unsubLog();
    }
  });

  function scrollTerminal() {
    if (terminalContainerRef) {
      terminalContainerRef.scrollTop = terminalContainerRef.scrollHeight;
    }
  }

  async function handleScanRepo() {
    if (!repoUrl.trim() || isScanning) return;
    isScanning = true;
    scanError = null;

    try {
      if (window.go?.main?.App?.ScanGitHubSkills) {
        const res = await window.go.main.App.ScanGitHubSkills(repoUrl.trim());
        analysisResult = res;
        // Default select all skills
        selectedSkillPaths = (res.skills || []).map((s) => s.relativePath);
        currentStep = 2;
      } else {
        // Fallback demo simulation if running in preview
        await new Promise((r) => setTimeout(r, 1200));
        analysisResult = {
          repoUrl: repoUrl.trim(),
          repoName: 'awesome-agent-skills',
          tempPath: '/tmp/aethergrok-skills/mock-skills',
          skills: [
            {
              name: 'browser-automator',
              description: 'Autonomous Chrome control via CDP protocol with click, type, and DOM query commands.',
              category: 'Agents',
              tags: ['browser', 'cdp', 'automation'],
              relativePath: 'skills/browser-automator',
              skillFile: 'skills/browser-automator/SKILL.md',
              prereqs: ['Node.js (npm)', 'Playwright'],
              commands: ['npm install playwright']
            },
            {
              name: 'fast-diff-reviewer',
              description: 'AI code review tool for git staging trees with precise line annotations.',
              category: 'Tools',
              tags: ['git', 'review', 'diff'],
              relativePath: 'skills/fast-diff-reviewer',
              skillFile: 'skills/fast-diff-reviewer/SKILL.md',
              prereqs: ['Python 3 (pip)'],
              commands: ['pip install gitpython']
            }
          ],
          globalPrereqs: ['Node.js (npm)', 'Python 3 (pip)'],
          suggestedScripts: ['npm install', 'pip install -r requirements.txt']
        };
        selectedSkillPaths = analysisResult.skills.map((s) => s.relativePath);
        currentStep = 2;
      }
    } catch (err) {
      scanError = String(err);
    } finally {
      isScanning = false;
    }
  }

  function toggleSkillSelection(relPath: string) {
    if (selectedSkillPaths.includes(relPath)) {
      selectedSkillPaths = selectedSkillPaths.filter((p) => p !== relPath);
    } else {
      selectedSkillPaths = [...selectedSkillPaths, relPath];
    }
  }

  function toggleSelectAll() {
    if (!analysisResult) return;
    if (selectedSkillPaths.length === analysisResult.skills.length) {
      selectedSkillPaths = [];
    } else {
      selectedSkillPaths = analysisResult.skills.map((s) => s.relativePath);
    }
  }

  async function handleStartInstall() {
    if (!analysisResult || selectedSkillPaths.length === 0 || isInstalling) return;
    currentStep = 3;
    isInstalling = true;
    installLogs = [];
    installResult = null;

    installLogs.push(`[info] Starting installation of ${selectedSkillPaths.length} skills to ~/.${targetScope}/skills/...`);

    try {
      if (window.go?.main?.App?.InstallDiscoveredSkills) {
        const payload: SkillInstallPayload = {
          tempPath: analysisResult.tempPath,
          skillPaths: selectedSkillPaths,
          targetScope
        };

        const res = await window.go.main.App.InstallDiscoveredSkills(payload);
        installResult = res;

        for (const path of res.installedPaths || []) {
          installLogs.push(`[success] Installed: ${path}`);
        }

        // Run prerequisite setup commands if chosen and available
        if (runSuggestedScripts && analysisResult.suggestedScripts.length > 0 && window.go?.main?.App?.ExecuteSkillSetupCommand) {
          installLogs.push(`[info] Executing ${analysisResult.suggestedScripts.length} dependency installation script(s)...`);
          for (const cmd of analysisResult.suggestedScripts) {
            installLogs.push(`[cmd] $ ${cmd}`);
            try {
              await window.go.main.App.ExecuteSkillSetupCommand(analysisResult.tempPath, cmd);
            } catch (cmdErr) {
              installLogs.push(`[warn] Command exited with: ${String(cmdErr)}`);
            }
          }
        }

        installLogs.push(`[complete] Successfully finished installing ${res.installedCount} skill(s).`);
      } else {
        // Mock installation execution
        await new Promise((r) => setTimeout(r, 600));
        installLogs.push(`[success] Installed to ~/.${targetScope}/skills/browser-automator`);
        installLogs.push(`[cmd] $ npm install`);
        installLogs.push(`[log] added 45 packages in 1.2s`);
        installLogs.push(`[complete] Done.`);
        installResult = {
          success: true,
          installedCount: selectedSkillPaths.length,
          installedPaths: selectedSkillPaths.map((p) => `~/.${targetScope}/skills/${p}`)
        };
      }

      onInstalled();
    } catch (err) {
      installLogs.push(`[error] Installation failed: ${String(err)}`);
    } finally {
      isInstalling = false;
      // Cleanup temp files
      if (analysisResult?.tempPath && window.go?.main?.App?.CleanupSkillImportTemp) {
        window.go.main.App.CleanupSkillImportTemp(analysisResult.tempPath).catch(() => {});
      }
    }
  }

  function resetState() {
    currentStep = 1;
    repoUrl = '';
    analysisResult = null;
    selectedSkillPaths = [];
    installLogs = [];
    installResult = null;
    scanError = null;
    isScanning = false;
    isInstalling = false;
  }

  function handleCloseModal() {
    if (analysisResult?.tempPath && window.go?.main?.App?.CleanupSkillImportTemp) {
      window.go.main.App.CleanupSkillImportTemp(analysisResult.tempPath).catch(() => {});
    }
    resetState();
    onClose();
  }
</script>

{#if visible}
  <div
    class="fixed inset-0 z-[110] flex items-center justify-center bg-black/75 backdrop-blur-md p-4 select-none animate-in fade-in duration-150"
    role="presentation"
    onclick={(e) => { if (e.target === e.currentTarget && !isScanning && !isInstalling) handleCloseModal(); }}
  >
    <div
      class="w-full max-w-3xl max-h-[85vh] flex flex-col bg-ant-bg-secondary border border-white/10 rounded-2xl shadow-2xl overflow-hidden animate-in zoom-in-95 duration-150 text-ant-text"
      role="dialog"
      aria-modal="true"
      tabindex="-1"
    >
      <!-- Top Title Header -->
      <div class="px-6 py-4 border-b border-white/5 flex items-center justify-between bg-ant-bg-secondary flex-shrink-0">
        <div class="flex items-center space-x-3">
          <div class="w-8 h-8 rounded-lg bg-ant-primary/10 border border-ant-primary/20 flex items-center justify-center text-ant-primary">
            <FolderGit2 size={17} />
          </div>
          <div>
            <div class="flex items-center space-x-2">
              <h2 class="font-serif-display text-base font-semibold text-ant-text tracking-tight">Import Skills from GitHub</h2>
              <span class="px-2 py-0.5 text-[10px] font-mono bg-white/5 text-ant-text-muted rounded border border-white/5">
                Step {currentStep} of 3
              </span>
            </div>
            <p class="text-xs text-ant-text-muted mt-0.5">
              Clone repositories, analyze dependencies, and install <code class="text-ant-primary font-mono text-[11px]">SKILL.md</code> recipes directly.
            </p>
          </div>
        </div>

        <button
          type="button"
          disabled={isScanning || isInstalling}
          onclick={handleCloseModal}
          class="p-1.5 rounded-lg text-ant-text-muted hover:text-ant-text hover:bg-ant-bg-tertiary transition disabled:opacity-30"
          title="Close Importer"
        >
          <X size={17} />
        </button>
      </div>

      <!-- Step Content Switcher -->
      <div class="flex-1 overflow-y-auto p-6 scrollbar-thin">
        {#if currentStep === 1}
          <!-- STEP 1: Repository URL Input -->
          <div class="max-w-xl mx-auto space-y-6 py-4">
            <div class="space-y-2">
              <label for="repo-url-input" class="block text-xs font-semibold text-ant-text flex items-center gap-1.5">
                <GitBranch size={13} class="text-ant-primary" />
                <span>GitHub Repository URL or Slug</span>
              </label>
              <div class="relative">
                <input
                  id="repo-url-input"
                  bind:value={repoUrl}
                  onkeydown={(e) => e.key === 'Enter' && handleScanRepo()}
                  type="text"
                  placeholder="https://github.com/owner/repository or owner/repository"
                  class="w-full px-3.5 py-2.5 bg-ant-bg border border-white/10 focus:border-ant-primary/60 rounded-xl text-xs text-ant-text placeholder:text-ant-text-muted outline-none transition shadow-sm font-mono"
                />
              </div>
              <p class="text-[11px] text-ant-text-muted">
                Accepts full GitHub HTTPS URLs or simple <code class="text-ant-text bg-ant-bg px-1 py-0.2 rounded border border-white/5">username/repo</code> format.
              </p>
            </div>

            <!-- Quick Suggestions / Popular Skills Repos -->
            <div class="p-3.5 bg-ant-bg rounded-xl border border-white/5 space-y-2">
              <div class="text-[11px] font-semibold text-ant-text-secondary flex items-center gap-1">
                <Sparkles size={12} class="text-ant-warning" />
                <span>Sample Repositories with Skills:</span>
              </div>
              <div class="flex flex-wrap gap-1.5">
                <button
                  type="button"
                  onclick={() => repoUrl = 'https://github.com/anthropics/skills'}
                  class="text-[11px] font-mono px-2 py-1 bg-ant-bg-secondary hover:bg-ant-primary/10 hover:text-ant-primary text-ant-text-muted rounded-md border border-white/5 transition"
                >
                  anthropics/skills
                </button>
                <button
                  type="button"
                  onclick={() => repoUrl = 'https://github.com/modelcontextprotocol/servers'}
                  class="text-[11px] font-mono px-2 py-1 bg-ant-bg-secondary hover:bg-ant-primary/10 hover:text-ant-primary text-ant-text-muted rounded-md border border-white/5 transition"
                >
                  modelcontextprotocol/servers
                </button>
              </div>
            </div>

            {#if scanError}
              <div class="p-3 rounded-xl bg-ant-error/10 border border-ant-error/20 text-ant-error text-xs flex items-start gap-2">
                <AlertCircle size={15} class="flex-shrink-0 mt-0.5" />
                <div class="leading-relaxed">
                  <span class="font-semibold">Failed to scan repository:</span> {scanError}
                </div>
              </div>
            {/if}

            <div class="pt-2 flex justify-end">
              <Button
                type="primary"
                onclick={handleScanRepo}
                disabled={!repoUrl.trim() || isScanning}
                class="!px-4 !py-2 text-xs flex items-center gap-1.5"
              >
                {#if isScanning}
                  <Loader2 size={13} class="animate-spin" />
                  <span>Analyzing Repository...</span>
                {:else}
                  <span>Scan Repository</span>
                  <ChevronRight size={13} />
                {/if}
              </Button>
            </div>
          </div>

        {:else if currentStep === 2 && analysisResult}
          <!-- STEP 2: Skill Multi-Select & Dependency Plan -->
          <div class="space-y-5">
            <!-- Header Summary Card -->
            <div class="p-3.5 bg-ant-bg rounded-xl border border-white/5 flex items-center justify-between">
              <div class="flex items-center gap-2">
                <FolderGit2 size={15} class="text-ant-primary" />
                <span class="text-xs font-semibold text-ant-text">{analysisResult.repoName}</span>
                <span class="text-[10px] text-ant-text-muted font-mono truncate max-w-xs">{analysisResult.repoUrl}</span>
              </div>
              <div class="flex items-center gap-2">
                <span class="text-[11px] text-ant-text-secondary">
                  Found <strong class="text-ant-text">{analysisResult.skills.length}</strong> skills
                </span>
              </div>
            </div>

            <!-- Global Dependency & Prereqs AI Summary -->
            {#if analysisResult.globalPrereqs.length > 0 || analysisResult.suggestedScripts.length > 0}
              <div class="p-3.5 bg-ant-bg-secondary rounded-xl border border-ant-primary/20 space-y-2.5">
                <div class="flex items-center justify-between">
                  <div class="flex items-center gap-1.5 text-xs font-semibold text-ant-primary">
                    <Wrench size={13} />
                    <span>AI Detected Prerequisites & Runtime Setup</span>
                  </div>
                  <label class="flex items-center gap-1.5 text-[11px] text-ant-text-secondary cursor-pointer">
                    <input type="checkbox" bind:checked={runSuggestedScripts} class="rounded text-ant-primary focus:ring-0" />
                    <span>Execute dependency scripts after copy</span>
                  </label>
                </div>

                <div class="flex flex-wrap gap-1.5">
                  {#each analysisResult.globalPrereqs as prereq}
                    <span class="px-2 py-0.5 rounded text-[10px] bg-ant-bg text-ant-text-secondary border border-white/5 font-mono">
                      {prereq}
                    </span>
                  {/each}
                </div>

                {#if runSuggestedScripts && analysisResult.suggestedScripts.length > 0}
                  <div class="bg-ant-bg p-2 rounded-lg border border-white/5 text-[11px] font-mono text-ant-text-muted space-y-1">
                    {#each analysisResult.suggestedScripts as script}
                      <div class="flex items-center gap-1 text-emerald-400">
                        <Terminal size={11} class="text-ant-text-muted" />
                        <span>$ {script}</span>
                      </div>
                    {/each}
                  </div>
                {/if}
              </div>
            {/if}

            <!-- Target Scope Selection -->
            <div class="flex items-center justify-between px-1">
              <div class="flex items-center gap-3 text-xs">
                <span class="text-ant-text-secondary font-medium">Install destination:</span>
                <label class="inline-flex items-center gap-1.5 cursor-pointer text-ant-text">
                  <input type="radio" bind:group={targetScope} value="grok" class="text-ant-primary" />
                  <span class="font-mono text-[11px]">~/.grok/skills/</span>
                </label>
                <label class="inline-flex items-center gap-1.5 cursor-pointer text-ant-text">
                  <input type="radio" bind:group={targetScope} value="agents" class="text-ant-primary" />
                  <span class="font-mono text-[11px]">~/.agents/skills/</span>
                </label>
              </div>

              <button
                type="button"
                onclick={toggleSelectAll}
                class="text-xs text-ant-primary hover:underline flex items-center gap-1"
              >
                {#if selectedSkillPaths.length === analysisResult.skills.length}
                  <CheckSquare size={13} />
                  <span>Deselect All</span>
                {:else}
                  <Square size={13} />
                  <span>Select All ({analysisResult.skills.length})</span>
                {/if}
              </button>
            </div>

            <!-- Discovered Skills Checklist Table -->
            <div class="border border-white/10 rounded-xl overflow-hidden bg-ant-bg">
              <div class="divide-y divide-white/5 max-h-72 overflow-y-auto scrollbar-thin">
                {#each analysisResult.skills as skill}
                  {@const isChecked = selectedSkillPaths.includes(skill.relativePath)}
                  <!-- svelte-ignore a11y_click_events_have_key_events -->
                  <!-- svelte-ignore a11y_no_static_element_interactions -->
                  <div
                    onclick={() => toggleSkillSelection(skill.relativePath)}
                    class="p-3 flex items-start gap-3 hover:bg-ant-bg-secondary/70 transition cursor-pointer {isChecked ? 'bg-ant-primary/5' : ''}"
                  >
                    <div class="mt-0.5 text-ant-primary flex-shrink-0">
                      {#if isChecked}
                        <CheckSquare size={15} class="text-ant-primary" />
                      {:else}
                        <Square size={15} class="text-ant-text-muted" />
                      {/if}
                    </div>

                    <div class="flex-1 min-w-0 space-y-1">
                      <div class="flex items-center gap-2">
                        <span class="font-semibold text-xs text-ant-text font-mono">/{skill.name}</span>
                        <span class="px-1.5 py-0.2 rounded text-[9.5px] bg-ant-bg-tertiary text-ant-text-muted font-mono border border-white/5">
                          {skill.category}
                        </span>
                        <span class="text-[10px] text-ant-text-muted font-mono truncate">
                          {skill.relativePath || 'root'}
                        </span>
                      </div>
                      <p class="text-[11px] text-ant-text-secondary leading-relaxed">
                        {skill.description}
                      </p>
                      {#if skill.prereqs && skill.prereqs.length > 0}
                        <div class="flex items-center gap-1.5 pt-0.5">
                          <span class="text-[9px] text-ant-text-muted">Requires:</span>
                          {#each skill.prereqs as req}
                            <span class="text-[9px] font-mono px-1 py-0.1 rounded bg-ant-bg text-ant-text-muted border border-white/5">
                              {req}
                            </span>
                          {/each}
                        </div>
                      {/if}
                    </div>
                  </div>
                {/each}
              </div>
            </div>

            <!-- Footer Action Buttons -->
            <div class="flex items-center justify-between pt-2">
              <Button
                type="default"
                size="small"
                onclick={() => currentStep = 1}
              >
                Back
              </Button>

              <Button
                type="primary"
                size="small"
                disabled={selectedSkillPaths.length === 0}
                onclick={handleStartInstall}
                class="!px-4 !py-1.5 flex items-center gap-1.5"
              >
                <Download size={13} />
                <span>Install Selected ({selectedSkillPaths.length})</span>
              </Button>
            </div>
          </div>

        {:else if currentStep === 3}
          <!-- STEP 3: Installation & Real-Time Terminal Logs -->
          <div class="space-y-4">
            <div class="flex items-center justify-between">
              <div class="flex items-center gap-2">
                <Terminal size={15} class="text-ant-primary" />
                <span class="text-xs font-semibold text-ant-text">Installation Execution Log</span>
              </div>
              <div class="flex items-center gap-2">
                {#if isInstalling}
                  <Loader2 size={13} class="animate-spin text-ant-primary" />
                  <span class="text-xs text-ant-text-muted">Installing...</span>
                {:else if installResult?.success}
                  <CheckCircle2 size={14} class="text-emerald-400" />
                  <span class="text-xs text-emerald-400 font-medium">Successfully Completed</span>
                {/if}
              </div>
            </div>

            <!-- Live Terminal View -->
            <div
              bind:this={terminalContainerRef}
              class="w-full h-72 bg-ant-bg rounded-xl border border-white/10 p-3 font-mono text-[11px] leading-relaxed overflow-y-auto scrollbar-thin text-ant-text space-y-1 select-text"
            >
              {#each installLogs as log}
                {#if log.startsWith('[error]')}
                  <div class="text-rose-400">{log}</div>
                {:else if log.startsWith('[success]')}
                  <div class="text-emerald-400">{log}</div>
                {:else if log.startsWith('[cmd]')}
                  <div class="text-cyan-400 font-semibold">{log}</div>
                {:else if log.startsWith('[warn]')}
                  <div class="text-amber-400">{log}</div>
                {:else}
                  <div class="text-ant-text-secondary">{log}</div>
                {/if}
              {/each}
              {#if isInstalling}
                <div class="flex items-center gap-1 text-ant-text-muted">
                  <span class="animate-pulse">▍</span>
                </div>
              {/if}
            </div>

            <!-- Action Buttons -->
            <div class="flex justify-end pt-2">
              {#if !isInstalling}
                <Button
                  type="primary"
                  onclick={handleCloseModal}
                  class="!px-4 !py-1.5 text-xs"
                >
                  Done & Refresh Skills
                </Button>
              {/if}
            </div>
          </div>
        {/if}
      </div>
    </div>
  </div>
{/if}
