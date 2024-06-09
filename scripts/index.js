import puppeteer from "puppeteer";
import axios from "axios";

(async () => {
  const { argv } = process;
  if (argv.length < 3) {
    throw new Error("not enough args. You need to provide an URL to listen to.");
  }

  const url = argv[2];
  const browser = await puppeteer.launch({ headless: true, args: ["--no-sandbox"] });
  const page = await browser.newPage();
  await page.setRequestInterception(true);

  let listen = false;
  setTimeout(() => (listen = true), 15000);
  page.on("request", async (request) => {
    if (listen && request.url().endsWith(".m3u8")) {
      axios.post("http://localhost:8080/segment", {
        status: "success",
        message: request.url(),
      });
    }
    request.continue();
  });

  await page.goto(url);
})();

// (async () => {
//   const p = "/mnt/c/Users/kosta/OneDrive/Desktop/imgs";
//   const dir = fs.readdirSync(p);
//   const tsFiles = dir.filter((x) => x.endsWith(".ts"));
//   console.log(tsFiles);
//   let idx = 0;

//   const sendFile = () => {
//     const f = tsFiles[0];
//     const content = fs.readFileSync(path.join(p, f));
//     axios.post("http://localhost:8080/segment", {
//       data: content[0].toString("base64"),
//     });
//     idx += 1;
//     setTimeout(sendFile, 2000);
//   };

//   sendFile();
// })();
