import puppeteer from "puppeteer";

(async () => {
  try {
    const { argv } = process;
    if (argv.length < 2) {
      throw new Error("not enough args. You need to provide an twitch URL.");
    }
    const url = argv[2];
    const browser = await puppeteer.launch({
      headless: true,
      args: ["--no-sandbox"],
      executablePath: "/usr/bin/google-chrome",
    });
    const page = await browser.newPage();
    await page.setRequestInterception(true);
    let m3u8URL = "";

    page.on("request", async (request) => {
      if (m3u8URL !== "") {
        // console.log("FOUND REQUEST", m3u8URL);
        console.log(m3u8URL);
        process.exit();
      }
      const url = request.url();
      if (url.endsWith(".m3u8")) {
        m3u8URL = url;
      }
      request.continue();
    });
    await page.goto(url);
  } catch (error) {
    console.log("ERROR: ", error);
  }
})();
