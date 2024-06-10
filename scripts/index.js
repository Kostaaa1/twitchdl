import puppeteer from "puppeteer";
import axios from "axios";

const selectors = {
  settings:
    "#channel-player > div > div.Layout-sc-1xcs6mc-0.kqyuWK.player-controls__right-control-group > div:nth-child(1) > div:nth-child(2) > div > button",
  quality:
    "body > div.ScReactModalBase-sc-26ijes-0.jnlrwf.tw-dialog-layer > div > div > div > div > div.ScBalloonWrapper-sc-14jr088-0.fpYyAb.InjectLayout-sc-1i43xsx-0.fUcjXo.tw-balloon > div > div > div:nth-child(2) > div:nth-child(3)",
  quality1080p:
    "body > div.ScReactModalBase-sc-26ijes-0.jnlrwf.tw-dialog-layer > div > div > div > div > div.ScBalloonWrapper-sc-14jr088-0.fpYyAb.InjectLayout-sc-1i43xsx-0.fUcjXo.tw-balloon > div > div > div:nth-child(2) > div:nth-child(2)",
  startWatching:
    "#channel-player-gate > div > div > div.Layout-sc-1xcs6mc-0.idFuL > div > button",
};

(async () => {
  const { argv } = process;
  if (argv.length < 3) {
    throw new Error("not enough args. You need to provide an URL to listen to.");
  }
  const url = argv[2];

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
        axios.post("http://localhost:8080/segment", {
          status: "success",
          message: url,
        });
      } else {
        if (url.length === 10) urls = [];
      }
    }
    request.continue();
  });

  await page.goto(url, { waitUntil: "networkidle2" });

  const startWatching = await page.$(selectors.startWatching);
  if (startWatching) {
    await startWatching.click();
  }

  await page.waitForSelector(selectors.settings);
  await page.click(selectors.settings);

  await page.waitForSelector(selectors.quality);
  await page.click(selectors.quality);

  await page.waitForSelector(selectors.quality1080p);
  await page.click(selectors.quality1080p);

  shouldListen = true;
})();
