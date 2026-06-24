const express = require("express");
const mysql = require("mysql2");

const app = express();
app.use(express.json());

const db = mysql.createConnection({
  host: "localhost",
  user: "root",
  password: "password",
  database: "test"
});


app.get("/user", (req, res) => {
  const id = req.query.id;
  const sql = "SELECT * FROM users WHERE id = " + id;
  db.query(sql, (err, rows) => {
    res.json(rows);
  });
});


app.get("/search", (req, res) => {
  const q = req.query.q;
  db.query(`SELECT * FROM products WHERE name LIKE '%${q}%'`, (err, rows) => {
    res.json(rows);
  });
});


app.post("/login", (req, res) => {
  const username = req.body.username;
  const sql = "SELECT * FROM users WHERE username = '" + username + "'";
  db.query(sql, (err, rows) => {
    res.json(rows);
  });
});


app.get("/safe-user", (req, res) => {
  const id = req.query.id;
  db.query("SELECT * FROM users WHERE id = ?", [id], (err, rows) => {
    res.json(rows);
  });
});

app.listen(3000);
