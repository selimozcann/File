const express = require("express");
const axios = require("axios");
const http = require("http");
const https = require("https");
const got = require("got");

const app = express();
app.use(express.json());

app.get("/ssrf/axios-query", async (req, res) => {
  const url = req.query.url;
  const response = await axios.get(url);
  res.send(response.data);
});

app.post("/ssrf/fetch-body", async (req, res) => {
  const target = req.body.target;
  const response = await fetch(target);
  const text = await response.text();
  res.send(text);
});

app.get("/ssrf/http-get", (req, res) => {
  const target = req.query.target;

  http.get(target, (response) => {
    response.pipe(res);
  });
});

app.get("/ssrf/https-request", (req, res) => {
  const target = req.query.target;

  const request = https.request(target, (response) => {
    response.pipe(res);
  });

  request.end();
});

app.get("/ssrf/got", async (req, res) => {
  const target = req.query.target;
  const response = await got(target);
  res.send(response.body);
});

app.get("/ssrf/template", async (req, res) => {
  const host = req.query.host;
  const url = `https://${host}/api/status`;

  const response = await axios.get(url);
  res.send(response.data);
});

app.get("/ssrf/helper", async (req, res) => {
  const domain = req.query.domain;
  const url = buildUrl(domain);

  const response = await axios.get(url);
  res.send(response.data);
});

function buildUrl(domain) {
  return "https://" + domain + "/profile";
}

app.get("/ssrf/object", async (req, res) => {
  const input = {
    callback: req.body.callback
  };

  const response = await fetch(input.callback);
  const text = await response.text();
  res.send(text);
});

app.get("/ssrf/redirect-follow", async (req, res) => {
  const url = req.query.url;

  const response = await axios.get("https://trusted.example/fetch?url=" + url);
  res.send(response.data);
});

app.get("/safe/allowlist", async (req, res) => {
  const allowed = ["https://api.github.com", "https://example.com"];
  const url = req.query.url;

  if (!allowed.includes(url)) {
    return res.status(400).json({ error: "URL not allowed" });
  }

  const response = await axios.get(url);
  res.send(response.data);
});

app.get("/safe/static", async (req, res) => {
  const response = await axios.get("https://api.github.com");
  res.send(response.data);
});

app.listen(3000);
