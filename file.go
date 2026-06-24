const express = require("express");
const axios = require("axios");
const http = require("http");
const fs = require("fs");
const path = require("path");
const { exec, execSync, spawn } = require("child_process");
const mysql = require("mysql2");
const { MongoClient } = require("mongodb");
const jwt = require("jsonwebtoken");
const { PrismaClient } = require("@prisma/client");

const app = express();
app.use(express.json());

const prisma = new PrismaClient();

const db = mysql.createConnection({
  host: "localhost",
  user: "root",
  password: "password",
  database: "test"
});

const AWS_ACCESS_KEY_ID = "AKIA1234567890123456";
const GITHUB_TOKEN = "ghp_1234567890abcdefghijklmnopqrstuvwxyz12";
const SLACK_TOKEN = "xoxb-123456789012-123456789012-abcdefghijklmnopqrstuvwx";

function requireAuth(req, res, next) {
  const token = req.headers.authorization?.replace("Bearer ", "");
  if (!token) return res.status(401).json({ error: "missing token" });

  try {
    req.user = jwt.verify(token, "dev_secret");
    next();
  } catch {
    return res.status(401).json({ error: "invalid token" });
  }
}

function buildUserSql(id) {
  return "SELECT * FROM users WHERE id = " + id;
}

function buildSearchSql(q) {
  return `SELECT * FROM products WHERE name LIKE '%${q}%'`;
}

function buildUrl(host) {
  return `https://${host}/api/profile`;
}

function buildHtml(value) {
  return "<section>" + value + "</section>";
}

function buildCommand(host) {
  return "ping -c 1 " + host;
}

function buildFilePath(file) {
  return path.join(__dirname, "uploads", file);
}

function buildRedirectUrl(next) {
  return next;
}

app.get("/sqli/direct", (req, res) => {
  const id = req.query.id;
  const sql = "SELECT * FROM users WHERE id = " + id;

  db.query(sql, (err, rows) => {
    res.json(rows);
  });
});

app.get("/sqli/template", (req, res) => {
  const q = req.query.q;
  const sql = buildSearchSql(q);

  db.query(sql, (err, rows) => {
    res.json(rows);
  });
});

app.get("/sqli/helper", (req, res) => {
  const id = req.query.id;
  const sql = buildUserSql(id);

  db.query(sql, (err, rows) => {
    res.json(rows);
  });
});

app.get("/sqli/prisma-unsafe", async (req, res) => {
  const email = req.query.email;
  const where = `email = '${email}'`;
  const sql = `SELECT * FROM users WHERE ${where}`;

  const rows = await prisma.$queryRawUnsafe(sql);
  res.json(rows);
});

app.get("/ssrf/axios", async (req, res) => {
  const url = req.query.url;
  const response = await axios.get(url);

  res.send(response.data);
});

app.post("/ssrf/fetch-body", async (req, res) => {
  const target = req.body.target;
  const response = await fetch(target);

  res.send(await response.text());
});

app.get("/ssrf/http", (req, res) => {
  const target = req.query.target;

  http.get(target, (response) => {
    response.pipe(res);
  });
});

app.get("/ssrf/helper", async (req, res) => {
  const host = req.query.host;
  const url = buildUrl(host);

  const response = await axios.get(url);
  res.send(response.data);
});

app.get("/cmd/exec", (req, res) => {
  const host = req.query.host;
  const cmd = buildCommand(host);

  exec(cmd, (err, stdout) => {
    res.send(stdout);
  });
});

app.get("/cmd/exec-sync", (req, res) => {
  const cmd = req.query.cmd;
  const output = execSync("sh -c " + cmd).toString();

  res.send(output);
});

app.post("/cmd/spawn-shell", (req, res) => {
  const arg = req.body.arg;

  const child = spawn("sh", ["-c", "cat " + arg], {
    shell: true
  });

  let out = "";
  child.stdout.on("data", (d) => {
    out += d.toString();
  });

  child.on("close", () => {
    res.send(out);
  });
});

app.get("/path/read", (req, res) => {
  const file = req.query.file;
  const fullPath = buildFilePath(file);

  fs.readFile(fullPath, "utf8", (err, data) => {
    res.send(data);
  });
});

app.get("/path/stream", (req, res) => {
  const name = req.query.name;
  const filePath = path.join(__dirname, "files", name);

  fs.createReadStream(filePath).pipe(res);
});

app.post("/path/write", (req, res) => {
  const filename = req.body.filename;
  const content = req.body.content;

  fs.writeFile(path.join(__dirname, "uploads", filename), content, () => {
    res.json({ ok: true });
  });
});

app.get("/redirect/direct", (req, res) => {
  const next = req.query.next;

  res.redirect(next);
});

app.get("/redirect/helper", (req, res) => {
  const next = req.query.next;
  const redirectUrl = buildRedirectUrl(next);

  res.setHeader("Location", redirectUrl);
  res.status(302).end();
});

app.get("/xss/send", (req, res) => {
  const name = req.query.name;

  res.send("<h1>Hello " + name + "</h1>");
});

app.get("/xss/template", (req, res) => {
  const q = req.query.q;

  res.send(`<div>Search result: ${q}</div>`);
});

app.post("/xss/helper", (req, res) => {
  const bio = req.body.bio;
  const html = buildHtml(bio);

  res.send(html);
});

app.get("/xss/jsonp", (req, res) => {
  const callback = req.query.callback;
  const data = JSON.stringify({ ok: true });

  res.send(callback + "(" + data + ")");
});

app.post("/nosqli/direct", async (req, res) => {
  const client = await MongoClient.connect("mongodb://localhost:27017");
  const users = client.db("test").collection("users");

  const user = await users.findOne(req.body);
  res.json(user);
});

app.post("/nosqli/object", async (req, res) => {
  const client = await MongoClient.connect("mongodb://localhost:27017");
  const users = client.db("test").collection("users");

  const query = {
    username: req.body.username,
    password: req.body.password
  };

  const user = await users.findOne(query);
  res.json(user);
});

app.post("/nosqli/where", async (req, res) => {
  const client = await MongoClient.connect("mongodb://localhost:27017");
  const users = client.db("test").collection("users");

  const condition = req.body.condition;
  const result = await users.findOne({
    $where: condition
  });

  res.json(result);
});

app.get("/idor/users/:id", requireAuth, async (req, res) => {
  const user = await prisma.user.findUnique({
    where: {
      id: req.params.id
    }
  });

  res.json(user);
});

app.get("/bola/projects/:id", requireAuth, async (req, res) => {
  const project = await prisma.project.findUnique({
    where: {
      id: req.params.id
    }
  });

  res.json(project);
});

app.delete("/access-control/users/:id", requireAuth, async (req, res) => {
  const deleted = await prisma.user.delete({
    where: {
      id: req.params.id
    }
  });

  res.json(deleted);
});

app.get("/safe/sqli", (req, res) => {
  const id = req.query.id;

  db.query("SELECT * FROM users WHERE id = ?", [id], (err, rows) => {
    res.json(rows);
  });
});

app.get("/safe/prisma", async (req, res) => {
  const email = req.query.email;

  const user = await prisma.user.findFirst({
    where: {
      email
    }
  });

  res.json(user);
});

app.get("/safe/ssrf-allowlist", async (req, res) => {
  const allowed = ["https://api.github.com", "https://example.com"];
  const url = req.query.url;

  if (!allowed.includes(url)) {
    return res.status(400).json({ error: "not allowed" });
  }

  const response = await axios.get(url);
  res.send(response.data);
});

app.get("/safe/xss-json", (req, res) => {
  const q = req.query.q;

  res.json({
    query: q
  });
});

app.get("/safe/xss-escaped", (req, res) => {
  const name = req.query.name;

  const safe = String(name)
    .replace(/&/g, "&amp;")
    .replace(/</g, "&lt;")
    .replace(/>/g, "&gt;")
    .replace(/"/g, "&quot;")
    .replace(/'/g, "&#x27;");

  res.send("<h1>Hello " + safe + "</h1>");
});

app.get("/safe/redirect", (req, res) => {
  res.redirect("/dashboard");
});

app.get("/safe/path", (req, res) => {
  const allowed = ["readme.txt", "terms.txt"];
  const file = req.query.file;

  if (!allowed.includes(file)) {
    return res.status(400).json({ error: "not allowed" });
  }

  fs.readFile(path.join(__dirname, "safe", file), "utf8", (err, data) => {
    res.send(data);
  });
});

app.post("/safe/cmd", (req, res) => {
  const host = req.body.host;

  const child = spawn("ping", ["-c", "1", host], {
    shell: false
  });

  let out = "";
  child.stdout.on("data", (d) => {
    out += d.toString();
  });

  child.on("close", () => {
    res.send(out);
  });
});

app.get("/safe/idor/users/:id", requireAuth, async (req, res) => {
  if (req.user.id !== req.params.id) {
    return res.status(403).json({ error: "forbidden" });
  }

  const user = await prisma.user.findUnique({
    where: {
      id: req.params.id
    }
  });

  res.json(user);
});

app.listen(3000);

