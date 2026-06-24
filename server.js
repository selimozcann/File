const express = require("express");

const app = express();
app.use(express.json());

app.get("/xss/query-send", (req, res) => {
  const name = req.query.name;
  res.send("<h1>Hello " + name + "</h1>");
});

app.get("/xss/template-send", (req, res) => {
  const q = req.query.q;
  res.send(`<div>Search result for: ${q}</div>`);
});

app.post("/xss/body-html", (req, res) => {
  const bio = req.body.bio;
  const html = "<html><body><p>" + bio + "</p></body></html>";
  res.send(html);
});

app.get("/xss/header", (req, res) => {
  const title = req.query.title;
  res.setHeader("Content-Type", "text/html");
  res.end("<title>" + title + "</title>");
});

app.get("/xss/render", (req, res) => {
  const username = req.query.username;
  res.render("profile", {
    html: username
  });
});

app.get("/xss/jsonp", (req, res) => {
  const callback = req.query.callback;
  const data = JSON.stringify({ ok: true });
  res.send(callback + "(" + data + ")");
});

app.get("/xss/redirect-fragment", (req, res) => {
  const message = req.query.message;
  res.send(`
    <script>
      document.body.innerHTML = "${message}";
    </script>
  `);
});

app.get("/xss/helper", (req, res) => {
  const value = req.query.value;
  const html = buildHtml(value);
  res.send(html);
});

function buildHtml(value) {
  return "<section>" + value + "</section>";
}

app.get("/safe/text", (req, res) => {
  const name = req.query.name;
  res.type("text/plain");
  res.send("Hello " + name);
});

app.get("/safe/escape", (req, res) => {
  const name = req.query.name;
  const safe = String(name)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#x27;");

  res.send("<h1>Hello " + safe + "</h1>");
});

app.get("/safe/json", (req, res) => {
  const q = req.query.q;
  res.json({
    query: q
  });
});

app.listen(3000);
