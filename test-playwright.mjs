import { chromium } from 'playwright';
(async () => {
  try {
    const browser = await chromium.launch({ headless: true });
    console.log("Success headless");
    await browser.close();
  } catch(e) { console.error("Error headless:", e); }
})();
