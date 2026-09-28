<script lang="ts">
  import Navbar from './lib/components/Navbar.svelte';
  import Hero from './lib/components/Hero.svelte';
  import ValueProps from './lib/components/ValueProps.svelte';
  import FeatureShowcase from './lib/components/FeatureShowcase.svelte';
  import ComparisonSection from './lib/components/ComparisonSection.svelte';
  import DownloadSection from './lib/components/DownloadSection.svelte';
  import MacGatekeeperGuide from './lib/components/MacGatekeeperGuide.svelte';
  import DonateSection from './lib/components/DonateSection.svelte';
  import Footer from './lib/components/Footer.svelte';
</script>

<div class="min-h-screen flex flex-col bg-white text-slate-900 selection:bg-blue-100 selection:text-blue-900">
  <Navbar />
  <main class="flex-grow">
    <Hero />
    <ValueProps />
    <FeatureShowcase />
    <ComparisonSection />
    <DownloadSection />
    <MacGatekeeperGuide />
    <DonateSection />
  </main>
  <Footer />
</div>
