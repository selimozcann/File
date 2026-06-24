const express = require("express");
const axios = require("axios");
const http = require("http");
const https = require("https");

const app = express();
app.use(express.json());

app.get("/fetch", async (req, res) => {
  const url = req.query.url;
  const response = await axios.get(url);
  res.send(response.data);
});

app.post("/webhook/test", async (req, res) => {
  const target = req.body.target;
  const response = await fetch(target);
  const text = await response.text();
  res.send(text);
});

app.get("/proxy", (req, res) => {
  const target = req.query.target;
  http.get(target, (response) => {
    response.pipe(res);
  });
});

app.get("/secure-fetch", async (req, res) => {
  const allowed = ["https://api.github.com", "https://example.com"];
  const url = req.query.url;

  if (!allowed.includes(url)) {
    return res.status(400).json({ error: "URL not allowed" });
  }

  const response = await axios.get(url);
  res.send(response.data);
});

app.listen(3000);
