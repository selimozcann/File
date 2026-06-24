const express = require("express");
const mysql = require("mysql2/promise");
const { PrismaClient } = require("@prisma/client");

const app = express();
app.use(express.json());

const prisma = new PrismaClient();

async function getDb() {
  return mysql.createConnection({
    host: "localhost",
    user: "root",
    password: "password",
    database: "test"
  });
}

// 1. Cross-variable taint
app.get("/advanced/user", async (req, res) => {
  const userId = req.query.id;
  const condition = "id = " + userId;
  const sql = "SELECT * FROM users WHERE " + condition;

  const db = await getDb();
  const [rows] = await db.query(sql);
  res.json(rows);
});

// 2. Helper function query builder
function buildUserQuery(id) {
  return "SELECT * FROM users WHERE id = " + id;
}

app.get("/advanced/helper", async (req, res) => {
  const id = req.query.id;
  const sql = buildUserQuery(id);

  const db = await getDb();
  const [rows] = await db.query(sql);
  res.json(rows);
});

// 3. Object property taint
app.get("/advanced/object", async (req, res) => {
  const filters = {
    username: req.query.username
  };

  const sql = "SELECT * FROM users WHERE username = '" + filters.username + "'";

  const db = await getDb();
  const [rows] = await db.query(sql);
  res.json(rows);
});

// 4. Array join query
app.get("/advanced/join", async (req, res) => {
  const id = req.params.id;
  const parts = [
    "SELECT * FROM orders WHERE user_id = ",
    id
  ];

  const sql = parts.join("");

  const db = await getDb();
  const [rows] = await db.query(sql);
  res.json(rows);
});

// 5. Prisma unsafe with variable
app.get("/advanced/prisma", async (req, res) => {
  const email = req.query.email;
  const where = `email = '${email}'`;
  const sql = `SELECT * FROM users WHERE ${where}`;

  const rows = await prisma.$queryRawUnsafe(sql);
  res.json(rows);
});

// 6. Sequelize-like raw query
app.get("/advanced/sequelize", async (req, res) => {
  const sort = req.query.sort || "id";
  const sql = `SELECT * FROM users ORDER BY ${sort}`;

  await sequelize.query(sql);
  res.json({ ok: true });
});

// 7. Safe parameterized mysql query - should NOT be flagged
app.get("/advanced/safe-param", async (req, res) => {
  const id = req.query.id;

  const db = await getDb();
  const [rows] = await db.query("SELECT * FROM users WHERE id = ?", [id]);
  res.json(rows);
});

// 8. Safe Prisma query builder - should NOT be flagged
app.get("/advanced/safe-prisma", async (req, res) => {
  const email = req.query.email;

  const user = await prisma.user.findFirst({
    where: { email }
  });

  res.json(user);
});

app.listen(3000);
