import puppeteer from "puppeteer";
import axios from "axios";

const selectors = {
  startWatching:
    "#channel-player-gate > div > div > div.Layout-sc-1xcs6mc-0.idFuL > div > button > div > div",
  settings:
    "#channel-player > div > div.Layout-sc-1xcs6mc-0.kqyuWK.player-controls__right-control-group > div:nth-child(1) > div:nth-child(2) > div > button",
  quality:
    "body > div.ScReactModalBase-sc-26ijes-0.kKKebR.tw-dialog-layer > div > div > div > div > div.ScBalloonWrapper-sc-14jr088-0.fpYyAb.InjectLayout-sc-1i43xsx-0.fUcjXo.tw-balloon > div > div > div:nth-child(2) > div:nth-child(3)",
  quality1080p:
    "body > div.ScReactModalBase-sc-26ijes-0.kKKebR.tw-dialog-layer > div > div > div > div > div.ScBalloonWrapper-sc-14jr088-0.fpYyAb.InjectLayout-sc-1i43xsx-0.fUcjXo.tw-balloon > div > div > div:nth-child(2) > div:nth-child(2)",
};

(async () => {
  try {
    const { argv } = process;
    if (argv.length < 3) {
      throw new Error("not enough args. You need to provide an twitch URL and PORT to.");
    }
    const url = argv[2];
    const port = argv[3];

    const browser = await puppeteer.launch({
      headless: true,
      args: ["--no-sandbox"],
      executablePath: "/usr/bin/google-chrome",
    });
    const page = await browser.newPage();
    await page.setRequestInterception(true);

    let shouldListen = false;
    let urls = [];

    page.on("request", async (request) => {
      const url = request.url();
      if (shouldListen && url.endsWith(".ts")) {
        if (!urls.includes(url)) {
          urls.push(url);
          axios.post(`http://localhost:${port}/segment`, {
            status: "success",
            message: url,
          });
        } else {
          if (urls.length === 10) urls = [];
        }
      }
      request.continue();
    });
    await page.goto(url);

    const startWatching = await page.$(selectors.startWatching);
    if (startWatching) await startWatching.click();
    await page.waitForSelector(selectors.settings);
    console.log("awaited settings");
    await page.click(selectors.settings);
    console.log("clicked settings");
    await page.waitForSelector(selectors.quality);
    console.log("awaited quality");
    await page.click(selectors.quality);
    console.log("clicked quality");
    await page.waitForSelector(selectors.quality1080p);
    console.log("awaited 1080");
    await page.click(selectors.quality1080p);
    console.log("clicked 1080");

    shouldListen = true;
  } catch (error) {
    console.log("ERROR: ", error);
  }
})();
