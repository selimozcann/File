const express = require("express");
const axios = require("axios");
const http = require("http");
const fs = require("fs");
const path = require("path");
const { exec } = require("child_process");
const mysql = require("mysql2");
const { MongoClient } = require("mongodb");
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

app.get("/sqli/concat", (req, res) => {
  const id = req.query.id;
  const sql = "SELECT * FROM users WHERE id = " + id;

  db.query(sql, (err, rows) => {
    res.json(rows);
  });
});

app.get("/sqli/template", (req, res) => {
  const q = req.query.q;

  db.query(`SELECT * FROM products WHERE name LIKE '%${q}%'`, (err, rows) => {
    res.json(rows);
  });
});

app.get("/sqli/prisma", async (req, res) => {
  const email = req.query.email;
  const sql = `SELECT * FROM users WHERE email = '${email}'`;

  const rows = await prisma.$queryRawUnsafe(sql);
  res.json(rows);
});

app.get("/ssrf/axios", async (req, res) => {
  const url = req.query.url;
  const response = await axios.get(url);

  res.send(response.data);
});

app.get("/ssrf/http", (req, res) => {
  const target = req.query.target;

  http.get(target, (response) => {
    response.pipe(res);
  });
});

app.get("/cmd/ping", (req, res) => {
  const host = req.query.host;

  exec("ping -c 1 " + host, (err, stdout) => {
    res.send(stdout);
  });
});

app.get("/path/read", (req, res) => {
  const file = req.query.file;
  const fullPath = path.join(__dirname, "uploads", file);

  fs.readFile(fullPath, "utf8", (err, data) => {
    res.send(data);
  });
});

app.get("/redirect", (req, res) => {
  const next = req.query.next;

  res.redirect(next);
});

app.get("/xss/send", (req, res) => {
  const name = req.query.name;

  res.send(`<h1>Hello ${name}</h1>`);
});

app.post("/nosqli/login", async (req, res) => {
  const client = await MongoClient.connect("mongodb://localhost:27017");
  const users = client.db("test").collection("users");

  const user = await users.findOne({
    username: req.body.username,
    password: req.body.password
  });

  res.json(user);
});

app.get("/idor/users/:id", async (req, res) => {
  const user = await prisma.user.findUnique({
    where: {
      id: req.params.id
    }
  });

  res.json(user);
});

app.delete("/access-control/users/:id", async (req, res) => {
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

app.get("/safe/redirect", (req, res) => {
  res.redirect("/dashboard");
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

app.listen(3000);
