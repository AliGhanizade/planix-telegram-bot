// planix web api client. every request carries the bearer session token.
"use strict";

const TOKEN_KEY = "planix_token";
const USER_KEY = "planix_user";

function getToken() { return localStorage.getItem(TOKEN_KEY); }
function getUser() {
  try { return JSON.parse(localStorage.getItem(USER_KEY)); } catch { return null; }
}
function saveSession(token, user) {
  localStorage.setItem(TOKEN_KEY, token);
  localStorage.setItem(USER_KEY, JSON.stringify(user));
}
function setStoredUser(user) { localStorage.setItem(USER_KEY, JSON.stringify(user)); }
function clearSession() {
  localStorage.removeItem(TOKEN_KEY);
  localStorage.removeItem(USER_KEY);
}

// api performs a json request against the backend.
async function api(path, { method = "GET", body, raw = false } = {}) {
  const headers = {};
  const token = getToken();
  if (token) headers["Authorization"] = "Bearer " + token;
  if (body !== undefined) headers["Content-Type"] = "application/json";

  const res = await fetch("/api" + path, {
    method,
    headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
  });

  if (res.status === 401) {
    clearSession();
    showLogin();
    throw new Error("unauthorized");
  }
  if (!res.ok) {
    let msg = "خطای نامشخص";
    try { msg = (await res.json()).error || msg; } catch { /* ignore */ }
    throw new Error(msg);
  }
  if (res.status === 204) return null;
  if (raw) return res;
  return res.json();
}

// login asks the bot to dm an otp, then exchanges the code for a session.
async function requestLoginCode(identifier) {
  return api("/auth/login/request", { method: "POST", body: { identifier } });
}

async function verifyLoginCode(code) {
  const data = await api("/auth/login/verify", { method: "POST", body: { code } });
  saveSession(data.token, data.user);
  return data;
}

// proofImageUrl returns a blob url for a task proof photo.
// the photo is streamed from telegram; the server never stores it.
async function fetchProofBlob(taskID) {
  const res = await api("/tasks/" + taskID + "/proof", { raw: true });
  return URL.createObjectURL(await res.blob());
}
