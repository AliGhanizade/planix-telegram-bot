// planix web panel: login, tasks, folders, profile and appearance customization.
"use strict";

// ---------- tiny helpers ----------

const $ = (sel) => document.querySelector(sel);
const $$ = (sel) => document.querySelectorAll(sel);

const icons = {
  "check-square": '<path d="m9 11 3 3L22 4"/><path d="M21 12v7a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11"/>',
  user: '<path d="M19 21v-2a4 4 0 0 0-4-4H9a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>',
  palette: '<circle cx="13.5" cy="6.5" r=".5" fill="currentColor"/><circle cx="17.5" cy="10.5" r=".5" fill="currentColor"/><circle cx="8.5" cy="7.5" r=".5" fill="currentColor"/><circle cx="6.5" cy="12.5" r=".5" fill="currentColor"/><path d="M12 2C6.5 2 2 6.5 2 12s4.5 10 10 10c.926 0 1.648-.746 1.648-1.688 0-.437-.18-.835-.437-1.125-.29-.289-.438-.652-.438-1.125a1.64 1.64 0 0 1 1.668-1.668h1.996c3.051 0 5.555-2.503 5.555-5.554C21.965 6.012 17.461 2 12 2z"/>',
  sun: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2"/><path d="M12 20v2"/><path d="m4.93 4.93 1.41 1.41"/><path d="m17.66 17.66 1.41 1.41"/><path d="M2 12h2"/><path d="M20 12h2"/><path d="m6.34 17.66-1.41 1.41"/><path d="m19.07 4.93-1.41 1.41"/>',
  moon: '<path d="M12 3a6 6 0 0 0 9 9 9 9 0 1 1-9-9Z"/>',
  logout: '<path d="M9 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h4"/><polyline points="16 17 21 12 16 7"/><line x1="21" x2="9" y1="12" y2="12"/>',
  close: '<path d="M18 6 6 18"/><path d="m6 6 12 12"/>',
  help: '<circle cx="12" cy="12" r="10"/><path d="M9.09 9a3 3 0 0 1 5.83 1c0 2-3 3-3 3"/><path d="M12 17h.01"/>',
};

function icon(name) {
  return '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">' + (icons[name] || "") + "</svg>";
}

function toast(msg, kind = "") {
  const el = document.createElement("div");
  el.className = "toast " + kind;
  el.textContent = msg;
  $("#toasts").appendChild(el);
  setTimeout(() => el.remove(), 3500);
}

function esc(s) {
  return String(s ?? "").replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" }[c]));
}

const dueFmt = (d) => d ? new Date(d).toLocaleString("fa-IR", { month: "2-digit", day: "2-digit", hour: "2-digit", minute: "2-digit" }) : "ندارد";

const priorityFa = { low: "کم", normal: "معمولی", high: "زیاد", urgent: "فوری" };
const statusFa = { pending: "باز", completed: "انجام‌شده", cancelled: "لغوشده" };

// ---------- view switching ----------

function showView(name) {
  $$(".view").forEach((v) => (v.hidden = true));
  $("#view-" + name).hidden = false;
  $$(".nav-link").forEach((b) => b.classList.toggle("active", b.dataset.view === name));
  $("#nav").hidden = false;
}

function showLogin() {
  $("#nav").hidden = true;
  $$(".view").forEach((v) => (v.hidden = true));
  $("#view-login").hidden = false;
  $("#stepIdentifier").hidden = false;
  $("#stepCode").hidden = true;
}

// ---------- login ----------

$("#stepIdentifier").addEventListener("submit", async (e) => {
  e.preventDefault();
  const id = $("#identifier").value.trim();
  if (!id) return;
  const btn = $("#requestCodeBtn");
  btn.disabled = true;
  try {
    await requestLoginCode(id);
    $("#stepIdentifier").hidden = true;
    $("#stepCode").hidden = false;
    $("#code").focus();
  } catch (err) {
    toast(err.message, "err");
  } finally {
    btn.disabled = false;
  }
});

$("#backBtn").addEventListener("click", () => {
  $("#stepIdentifier").hidden = false;
  $("#stepCode").hidden = true;
});

$("#stepCode").addEventListener("submit", async (e) => {
  e.preventDefault();
  const code = $("#code").value.trim();
  if (code.length !== 6) return toast("کد باید ۶ رقم باشد", "err");
  const btn = $("#verifyBtn");
  btn.disabled = true;
  try {
    await verifyLoginCode(code);
    enterApp();
  } catch (err) {
    toast(err.message, "err");
  } finally {
    btn.disabled = false;
  }
});

function enterApp() {
  showView("tasks");
  loadFolders();
  loadStats();
  loadTasks();
}

// ---------- nav ----------

$$(".nav-link").forEach((b) => b.addEventListener("click", () => {
  showView(b.dataset.view);
  if (b.dataset.view === "tasks") { loadFolders(); loadStats(); loadTasks(); }
  if (b.dataset.view === "delegations") loadDelegations();
  if (b.dataset.view === "profile") loadProfile();
}));

$("#logoutBtn").innerHTML = icon("logout");
$("#logoutBtn").addEventListener("click", async () => {
  try { await api("/auth/logout", { method: "POST" }); } catch { /* ignore */ }
  clearSession();
  showLogin();
});

// ---------- tasks ----------

const taskState = { filter: "pending", time: "", folder: null, page: 1, pages: 1, q: "" };

$$("#filterTabs .tab").forEach((t) => t.addEventListener("click", () => {
  $$("#filterTabs .tab").forEach((x) => x.classList.remove("active"));
  t.classList.add("active");
  taskState.filter = t.dataset.filter;
  taskState.page = 1;
  taskState.folder = null;
  renderFolderBar();
  loadTasks();
}));

$$("#timeTabs .tab").forEach((t) => t.addEventListener("click", () => {
  $$("#timeTabs .tab").forEach((x) => x.classList.remove("active"));
  t.classList.add("active");
  taskState.time = t.dataset.filter;
  taskState.page = 1;
  taskState.folder = null;
  renderFolderBar();
  loadTasks();
}));

let searchTimer;
$("#searchBox").addEventListener("input", (e) => {
  clearTimeout(searchTimer);
  searchTimer = setTimeout(() => {
    taskState.q = e.target.value.trim();
    taskState.page = 1;
    loadTasks();
  }, 350);
});

$("#prevPage").addEventListener("click", () => { if (taskState.page > 1) { taskState.page--; loadTasks(); } });
$("#nextPage").addEventListener("click", () => { if (taskState.page < taskState.pages) { taskState.page++; loadTasks(); } });
$("#newTaskBtn").addEventListener("click", () => openTaskModal());

async function loadStats() {
  try {
    const s = await api("/stats");
    $("#statPending").textContent = s.pending_tasks;
    $("#statCompleted").textContent = s.completed_tasks;
    $("#statCancelled").textContent = s.cancelled_tasks;
    $("#statProgress").textContent = Math.round(s.progress_percent) + "%";
    $("#progressBar").style.width = s.progress_percent + "%";
  } catch { /* ignore */ }
}

async function loadTasks() {
  const list = $("#taskList");
  list.innerHTML = '<div class="empty">در حال بارگذاری…</div>';

  // folder view has its own loader
  if (taskState.folder) { loadFolderTasks(taskState.folder); return; }

  const qs = new URLSearchParams({ status: taskState.filter, page: taskState.page });
  if (taskState.time) qs.set("status", taskState.time);
  if (taskState.q) qs.set("q", taskState.q);

  let data;
  try {
    data = await api("/tasks?" + qs);
  } catch (err) {
    list.innerHTML = '<div class="empty">' + esc(err.message) + "</div>";
    return;
  }

  taskState.pages = data.pages;
  $("#pageInfo").textContent = data.page + " / " + data.pages;

  if (!data.items.length) {
    list.innerHTML = '<div class="empty">تسکی اینجا نیست 🙌</div>';
    return;
  }

  list.innerHTML = data.items.map((t) => {
    const delegated = t.owner_id !== t.assignee_id;
    return `
    <div class="task-card" data-id="${t.id}">
      <div class="task-body">
        <p class="task-title ${t.status !== "pending" ? "done" : ""}">${esc(t.title)}</p>
        ${t.description ? `<p class="task-desc">${esc(t.description)}</p>` : ""}
        <div class="task-meta">
          <span class="chip priority ${t.priority}">${priorityFa[t.priority] || t.priority}</span>
          <span class="chip status ${t.status}">${statusFa[t.status] || t.status}</span>
          <span class="chip">موعد: ${dueFmt(t.due_at)}</span>
          ${t.requires_evidence ? '<span class="chip">🖼 نیاز به مدرک</span>' : ""}
          ${t.has_proof ? '<span class="chip proof" data-act="proof">📷 مشاهده مدرک</span>' : ""}
          ${delegated ? `<span class="chip">👤 مالک: ${esc(t.owner_name)}</span>` : ""}
          ${delegated ? `<span class="chip">🛠 مجری: ${esc(t.assignee_name)}</span>` : ""}
        </div>
      </div>
      <div class="task-actions">
        ${t.status === "pending"
          ? '<button class="btn small primary" data-act="done">انجام شد</button>'
          : (t.requires_evidence ? '<button class="btn small ghost" data-act="reopen">بازگشایی</button>' : "")}
        <button class="btn small ghost" data-act="edit">ویرایش</button>
        <button class="btn small danger" data-act="delete">حذف</button>
      </div>
    </div>
  `;
  }).join("");
}

$("#taskList").addEventListener("click", async (e) => {
  const card = e.target.closest(".task-card");
  if (!card) return;
  const id = card.dataset.id;
  const act = e.target.closest("[data-act]")?.dataset.act;

  if (act === "done" || act === "reopen") {
    try {
      await api("/tasks/" + id + "/" + (act === "done" ? "complete" : "reopen"), { method: "POST" });
      toast(act === "done" ? "تسک انجام شد ✅" : "تسک بازگشایی شد", "ok");
      taskState.folder ? loadFolderTasks(taskState.folder) : loadTasks();
      loadStats();
    } catch (err) { toast(err.message, "err"); }
  }

  if (act === "delete") {
    openConfirm("تسک حذف شود؟ این کار برگشت‌پذیر نیست.", async () => {
      try {
        await api("/tasks/" + id, { method: "DELETE" });
        toast("حذف شد", "ok");
        taskState.folder ? loadFolderTasks(taskState.folder) : loadTasks();
        loadStats();
      } catch (err) { toast(err.message, "err"); }
    });
  }

  if (act === "edit") {
    const t = await api("/tasks/" + id);
    openTaskModal(t);
  }

  if (act === "proof") openProof(id);
});

// ---------- folders ----------

let myFolders = [];

async function loadFolders() {
  try {
    const data = await api("/folders");
    myFolders = data.items || [];
  } catch { myFolders = []; }
  renderFolderBar();
}

function renderFolderBar() {
  const bar = $("#folderBar");
  if (!bar) return;
  let html = '<button class="folder-chip ' + (taskState.folder ? "" : "active") + '" data-fid="">همه‌ی تسک‌ها</button>';
  for (const f of myFolders) {
    html += '<button class="folder-chip ' + (taskState.folder === f.id ? "active" : "") + '" data-fid="' + f.id + '">📁 ' + esc(f.name) +
      (f.kind === "shared" ? ' <span class="fx">دریافتی</span>' : "") +
      (taskState.folder === f.id ? ' <span class="fx" data-del="' + f.id + '">حذف پوشه</span>' : "") + '</button>';
  }
  html += '<button class="folder-chip add" id="addFolderBtn">＋ پوشه جدید</button>';
  bar.innerHTML = html;

  bar.querySelectorAll(".folder-chip[data-fid]").forEach((b) => b.addEventListener("click", (e) => {
    if (e.target.dataset.del) return;
    taskState.folder = b.dataset.fid || null;
    taskState.page = 1;
    renderFolderBar();
    taskState.folder ? loadFolderTasks(taskState.folder) : loadTasks();
  }));
  const del = bar.querySelector("[data-del]");
  if (del) del.addEventListener("click", (e) => {
    e.stopPropagation();
    const fid = del.dataset.del;
    openConfirm("پوشه حذف شود؟ تسک‌ها حذف نمی‌شوند و فقط از این پوشه خارج می‌شوند.", async () => {
      try {
        await api("/folders/" + fid, { method: "DELETE" });
        toast("پوشه حذف شد", "ok");
        taskState.folder = null;
        loadFolders(); loadTasks();
      } catch (err) { toast(err.message, "err"); }
    });
  });
  $("#addFolderBtn")?.addEventListener("click", () => openFolderCreate());
}

function openFolderCreate() {
  openModal(`
    <h3>پوشه‌ی جدید</h3>
    <form id="folderForm">
      <label>نام پوشه<input name="name" required maxlength="80"></label>
      <label>داخل پوشه‌ی والد (اختیاری — برای زیرپوشه)
        <select name="parent">
          <option value="">— ریشه —</option>
          ${myFolders.filter((f) => f.kind === "own").map((f) => `<option value="${f.id}">${esc(f.name)}</option>`).join("")}
        </select>
      </label>
      <div class="row">
        <button type="button" class="btn ghost" data-close>انصراف</button>
        <button type="submit" class="btn primary">ساخت</button>
      </div>
    </form>
  `);
  $("#folderForm").addEventListener("submit", async (e) => {
    e.preventDefault();
    const f = new FormData(e.target);
    try {
      await api("/folders", { method: "POST", body: { name: f.get("name"), parent_id: f.get("parent") } });
      toast("پوشه ساخته شد", "ok");
      closeModal(); loadFolders();
    } catch (err) { toast(err.message, "err"); }
  });
}

async function loadFolderTasks(folderID) {
  const list = $("#taskList");
  list.innerHTML = '<div class="empty">در حال بارگذاری…</div>';
  $("#pageInfo").textContent = "پوشه";
  let data;
  try {
    data = await api("/folders/" + folderID + "/tasks");
  } catch (err) {
    list.innerHTML = '<div class="empty">' + esc(err.message) + "</div>";
    return;
  }
  if (!data.items.length) { list.innerHTML = '<div class="empty">این پوشه خالی است.</div>'; return; }
  list.innerHTML = data.items.map((t) => `
    <div class="task-card" data-id="${t.id}">
      <div class="task-body">
        <p class="task-title ${t.status !== "pending" ? "done" : ""}">${esc(t.title)}</p>
        ${t.description ? `<p class="task-desc">${esc(t.description)}</p>` : ""}
        <div class="task-meta">
          <span class="chip status ${t.status}">${statusFa[t.status] || t.status}</span>
          ${t.has_proof ? '<span class="chip proof" data-act="proof">📷 مشاهده مدرک</span>' : ""}
        </div>
      </div>
      <div class="task-actions">
        ${t.status === "pending" ? '<button class="btn small primary" data-act="done">انجام شد</button>' : ""}
      </div>
    </div>
  `).join("");
}

// ---------- task create / edit modal ----------

function openTaskModal(task) {
  const isEdit = !!task;
  openModal(`
    <h3>${isEdit ? "ویرایش تسک" : "تسک جدید"}</h3>
    <form id="taskForm">
      <label>عنوان
        <input name="title" required maxlength="240" value="${esc(task?.title || "")}">
      </label>
      <label>توضیحات
        <textarea name="description" rows="3">${esc(task?.description || "")}</textarea>
      </label>
      <label>اولویت
        <select name="priority">
          <option value="low" ${task?.priority === "low" ? "selected" : ""}>کم</option>
          <option value="normal" ${!task || task.priority === "normal" ? "selected" : ""}>معمولی</option>
          <option value="high" ${task?.priority === "high" ? "selected" : ""}>زیاد</option>
          <option value="urgent" ${task?.priority === "urgent" ? "selected" : ""}>فوری</option>
        </select>
      </label>
      <label>موعد
        <input type="datetime-local" name="due" value="${toLocalInput(task?.due_at)}">
      </label>
      <label class="check-row">
        <input type="checkbox" name="requires_evidence" ${task?.requires_evidence ? "checked" : ""}>
        نیاز به ارسال عکس مدرک دارد (با تیک خوردن این گزینه، بعد از انجام تسک قابل بازگشایی است)
      </label>
      ${!isEdit ? `<label>واگذاری به (اختیاری — @username یا آیدی عددی)
        <input name="assignee" placeholder="خالی = برای خودت">
      </label>` : ""}
      <label>پوشه‌ها
        <div class="folder-picker">
        ${myFolders.filter((f) => f.kind === "own").map((f) => {
          const linked = (task?.folders || []).some((x) => x.id === f.id);
          return '<label class="check-row"><input type="checkbox" name="folder" value="' + f.id + '"' + (linked ? " checked" : "") + ">📁 " + esc(f.name) + "</label>";
        }).join("") || '<span style="color:var(--text-dim)">هنوز پوشه‌ای نداری</span>'}
        </div>
      </label>
      <div class="row">
        <button type="button" class="btn ghost" data-close>انصراف</button>
        <button type="submit" class="btn primary">ذخیره</button>
      </div>
    </form>
  `);

  $("#taskForm").addEventListener("submit", async (e) => {
    e.preventDefault();
    const f = new FormData(e.target);
    const due = f.get("due");
    const body = {
      title: f.get("title").trim(),
      description: f.get("description"),
      priority: f.get("priority"),
      requires_evidence: f.get("requires_evidence") === "on",
    };
    const pickedFolders = [...e.target.querySelectorAll('input[name="folder"]:checked')].map((x) => x.value);
    const assignee = f.get("assignee") ? f.get("assignee").trim() : "";

    if (isEdit) {
      body.status = f.get("status");
      body.due_at = due ? new Date(due).toISOString() : null;
      try {
        await api("/tasks/" + task.id, { method: "PATCH", body });
        await syncTaskFolders(task.id, pickedFolders, task.folders || []);
        toast("ذخیره شد", "ok");
        closeModal();
        taskState.folder ? loadFolderTasks(taskState.folder) : loadTasks();
        loadStats();
      } catch (err) { toast(err.message, "err"); }
    } else {
      if (due) body.due_at = new Date(due).toISOString();
      if (assignee) body.assignee = assignee;
      let created;
      try {
        created = await api("/tasks", { method: "POST", body });
        for (const fid of pickedFolders) {
          await api("/tasks/" + created.id + "/folders", { method: "POST", body: { folder_id: fid } });
        }
        toast(assignee ? "تسک واگذار شد ✅" : "تسک ثبت شد ✅", "ok");
        closeModal();
        taskState.filter = assignee ? "all" : "pending";
        $$("#filterTabs .tab").forEach((x) => x.classList.toggle("active", x.dataset.filter === taskState.filter));
        taskState.page = 1;
        loadFolders(); loadTasks(); loadStats();
      } catch (err) { toast(err.message, "err"); }
    }
  });
}

// syncTaskFolders links and unlinks folders to match the picked list.
async function syncTaskFolders(taskID, picked, current) {
  const currentIds = new Set(current.map((f) => f.id));
  const pickedSet = new Set(picked);
  for (const fid of picked) {
    if (!currentIds.has(fid)) await api("/tasks/" + taskID + "/folders", { method: "POST", body: { folder_id: fid } });
  }
  for (const f of current) {
    if (!pickedSet.has(f.id)) await api("/tasks/" + taskID + "/folders/" + f.id, { method: "DELETE" });
  }
}

function toLocalInput(iso) {
  if (!iso) return "";
  const d = new Date(iso);
  const pad = (n) => String(n).padStart(2, "0");
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}T${pad(d.getHours())}:${pad(d.getMinutes())}`;
}

// ---------- proof photo (streamed from telegram, never stored) ----------

async function openProof(id) {
  openModal('<h3>در حال دریافت مدرک…</h3>');
  try {
    const url = await fetchProofBlob(id);
    $("#modalBox").innerHTML = `
      <div class="modal-head-row" style="display:flex;justify-content:space-between;align-items:center"><h3 style="margin:0">مدرک تسک</h3><button class="icon-btn" data-close>${icon("close")}</button></div>
      <img class="proof" src="${url}" alt="proof">
    `;
    $$("[data-close]").forEach((b) => b.addEventListener("click", closeModal));
  } catch (err) {
    closeModal();
    toast(err.message, "err");
  }
}

// ---------- profile ----------

async function loadProfile() {
  const card = $("#profileCard");
  card.innerHTML = '<div class="empty">در حال بارگذاری…</div>';
  try {
    const p = await api("/profile");
    const username = p.username ? "@" + p.username : "ندارد";
    card.innerHTML = `
      <h3 style="margin:0 0 4px">${esc(p.first_name)} ${esc(p.last_name)}</h3>
      <p style="margin:0;color:var(--text-dim);font-size:.9rem">شناسه تلگرام: ${p.telegram_id}</p>
      <div class="profile-grid">
        <div class="p-item"><div class="k">یوزرنیم</div><div class="v">${esc(username)}</div></div>
        <div class="p-item"><div class="k">تایم‌زون</div><div class="v">${esc(p.timezone)}</div></div>
        <div class="p-item"><div class="k">تسک‌های باز</div><div class="v">${p.pending_tasks}</div></div>
        <div class="p-item"><div class="k">انجام‌شده</div><div class="v">${p.completed_tasks}</div></div>
        <div class="p-item"><div class="k">پیشرفت</div><div class="v">${Math.round(p.progress_percent)}%</div></div>
      </div>
      <div class="toggle-row">
        <span>گزارش روزانه (خلاصه‌ی تسک‌های باز)</span>
        <button class="btn small ${p.daily_report ? "primary" : "ghost"}" id="dailyToggle">${p.daily_report ? "روشن" : "خاموش"}</button>
      </div>
      <div class="toggle-row">
        <span>ساعت‌های گزارش (حداکثر ۳): <b>${(p.report_times || []).join(" ، ") || "—"}</b></span>
        <button class="btn small ghost" id="timesEditor">ویرایش ساعت‌ها</button>
      </div>
      <div class="p-actions">
        <button class="btn primary" id="editProfileBtn">ویرایش اطلاعات</button>
      </div>
    `;

    $("#editProfileBtn").addEventListener("click", () => openProfileEdit(p));
    $("#dailyToggle").addEventListener("click", async () => {
      try {
        const res = await api("/profile/daily-report", { method: "PUT", body: { enabled: !p.daily_report } });
        toast(res.daily_report ? "گزارش روزانه روشن شد" : "گزارش روزانه خاموش شد", "ok");
        loadProfile();
      } catch (err) { toast(err.message, "err"); }
    });
    $("#timesEditor").addEventListener("click", () => openTimesEditor(p.report_times || []));
  } catch (err) {
    card.innerHTML = '<div class="empty">' + esc(err.message) + "</div>";
  }
}

function openProfileEdit(p) {
  openModal(`
    <h3>ویرایش اطلاعات</h3>
    <form id="profileForm">
      <label>نام<input name="first_name" value="${esc(p.first_name)}" maxlength="64"></label>
      <label>نام خانوادگی<input name="last_name" value="${esc(p.last_name)}" maxlength="64"></label>
      <label>تایم‌زون (مثل Asia/Tehran)<input name="timezone" value="${esc(p.timezone)}"></label>
      <div class="row">
        <button type="button" class="btn ghost" data-close>انصراف</button>
        <button type="submit" class="btn primary">ذخیره</button>
      </div>
    </form>
  `);

  $("#profileForm").addEventListener("submit", async (e) => {
    e.preventDefault();
    const f = new FormData(e.target);
    const body = {};
    if (f.get("first_name").trim()) body.first_name = f.get("first_name").trim();
    body.last_name = f.get("last_name").trim();
    body.timezone = f.get("timezone").trim();
    try {
      await api("/profile", { method: "PATCH", body });
      toast("اطلاعات ذخیره شد", "ok");
      closeModal(); loadProfile();
    } catch (err) { toast(err.message, "err"); }
  });
}

// times editor modal: up to three HH:MM values
function openTimesEditor(times) {
  const rows = [0, 1, 2].map((i) =>
    '<input type="time" name="t' + i + '" value="' + esc(times[i] || "") + '">'
  ).join("");
  openModal(`
    <h3>ساعت‌های گزارش روزانه</h3>
    <form id="timesForm">
      <p style="margin:0 0 10px;color:var(--text-dim);font-size:.85rem">در این ساعت‌ها (به وقت تهران) خلاصه‌ی تسک‌های بازت را می‌گیری. خالی گذاشتن یعنی گزارش نمی‌خواهی.</p>
      ${rows}
      <div class="row">
        <button type="button" class="btn ghost" data-close>انصراف</button>
        <button type="submit" class="btn primary">ذخیره</button>
      </div>
    </form>
  `);
  $("#timesForm").addEventListener("submit", async (e) => {
    e.preventDefault();
    const f = new FormData(e.target);
    const times = [...f.values()].map((v) => v.trim()).filter(Boolean);
    try {
      await api("/profile/report-times", { method: "PUT", body: { times } });
      toast("ذخیره شد", "ok");
      closeModal(); loadProfile();
    } catch (err) { toast(err.message, "err"); }
  });
}

// ---------- modal ----------

function openModal(html) {
  $("#modalBox").innerHTML = html;
  $("#modalBackdrop").hidden = false;
  $$("[data-close]").forEach((b) => b.addEventListener("click", closeModal));
}
function closeModal() { $("#modalBackdrop").hidden = true; }
$("#modalBackdrop").addEventListener("click", (e) => { if (e.target.id === "modalBackdrop") closeModal(); });

function openConfirm(text, onYes) {
  openModal(`
    <h3>تایید</h3>
    <p>${esc(text)}</p>
    <div class="row">
      <button class="btn ghost" data-close>انصراف</button>
      <button class="btn danger" id="confirmYes">بله، انجام بده</button>
    </div>
  `);
  $("#confirmYes").addEventListener("click", () => { closeModal(); onYes(); });
}

// ---------- delegations (help desk) ----------

const delState = { status: "pending", page: 1, pages: 1, q: "" };

$$("#delTabs .tab").forEach((t) => t.addEventListener("click", () => {
  $$("#delTabs .tab").forEach((x) => x.classList.remove("active"));
  t.classList.add("active");
  delState.status = t.dataset.filter;
  delState.page = 1;
  loadDelegations();
}));

let delSearchTimer;
$("#delSearch").addEventListener("input", (e) => {
  clearTimeout(delSearchTimer);
  delSearchTimer = setTimeout(() => {
    delState.q = e.target.value.trim();
    delState.page = 1;
    loadDelegations();
  }, 350);
});

$("#delPrev").addEventListener("click", () => { if (delState.page > 1) { delState.page--; loadDelegations(); } });
$("#delNext").addEventListener("click", () => { if (delState.page < delState.pages) { delState.page++; loadDelegations(); } });

async function loadDelegations() {
  const list = $("#delList");
  list.innerHTML = '<div class="empty">در حال بارگذاری…</div>';
  const qs = new URLSearchParams({ status: delState.status, page: delState.page });
  if (delState.q) qs.set("q", delState.q);

  let data;
  try {
    data = await api("/delegations?" + qs);
  } catch (err) {
    list.innerHTML = '<div class="empty">' + esc(err.message) + "</div>";
    return;
  }

  delState.pages = data.pages;
  $("#delPageInfo").textContent = data.page + " / " + data.pages;

  // per person summary from the current page
  const byPerson = {};
  for (const t of data.items) {
    const key = t.assignee_name || ("@" + t.assignee_username) || t.assignee_telegram_id;
    byPerson[key] = byPerson[key] || { total: 0, done: 0 };
    byPerson[key].total++;
    if (t.status === "completed") byPerson[key].done++;
  }
  const summary = $("#delSummary");
  const entries = Object.entries(byPerson);
  if (entries.length) {
    summary.hidden = false;
    summary.innerHTML = entries.map(([name, v]) =>
      '<div class="stat"><span class="stat-num">' + v.done + " / " + v.total + '</span><span class="stat-label">' + esc(name) + "</span></div>"
    ).join("");
  } else {
    summary.hidden = true;
  }

  if (!data.items.length) {
    list.innerHTML = '<div class="empty">تسکی واگذار نکرده‌ای.</div>';
    return;
  }

  list.innerHTML = data.items.map((t) => `
    <div class="task-card" data-id="${t.id}">
      <div class="task-body">
        <p class="task-title ${t.status !== "pending" ? "done" : ""}">${esc(t.title)}</p>
        ${t.description ? `<p class="task-desc">${esc(t.description)}</p>` : ""}
        <div class="task-meta">
          <span class="chip">🛠 مجری: ${esc(t.assignee_name)} (@${esc(t.assignee_username)})</span>
          <span class="chip">آیدی: ${t.assignee_telegram_id}</span>
          <span class="chip priority ${t.priority}">${priorityFa[t.priority] || t.priority}</span>
          <span class="chip status ${t.status}">${statusFa[t.status] || t.status}</span>
          <span class="chip">موعد: ${dueFmt(t.due_at)}</span>
          ${t.requires_evidence ? '<span class="chip">🖼 نیاز به مدرک</span>' : ""}
          ${t.has_proof ? '<span class="chip proof" data-act="proof">📷 مشاهده مدرک</span>' : ""}
        </div>
      </div>
      <div class="task-actions">
        ${t.status === "completed" && t.requires_evidence ? '<button class="btn small ghost" data-act="reopen">بازگشایی</button>' : ""}
        <button class="btn small danger" data-act="delete">حذف</button>
      </div>
    </div>
  `).join("");
}

// delegations card actions (reopen, proof, delete)
$("#delList").addEventListener("click", async (e) => {
  const card = e.target.closest(".task-card");
  if (!card) return;
  const id = card.dataset.id;
  const act = e.target.closest("[data-act]")?.dataset.act;

  if (act === "reopen") {
    try {
      await api("/tasks/" + id + "/reopen", { method: "POST" });
      toast("تسک بازگشایی شد", "ok");
      loadDelegations();
    } catch (err) { toast(err.message, "err"); }
  }
  if (act === "delete") {
    openConfirm("تسک حذف شود؟ این کار برگشت‌پذیر نیست.", async () => {
      try {
        await api("/tasks/" + id, { method: "DELETE" });
        toast("حذف شد", "ok");
        loadDelegations();
      } catch (err) { toast(err.message, "err"); }
    });
  }
  if (act === "proof") openProof(id);
});

// ---------- appearance: theme, style, accent, radius ----------

const APP_KEY = "planix_appearance";
const defaultLook = { theme: "dark", style: "matte", accent: "#3b82f6", radius: 14, btnRadius: 10 };
let look = { ...defaultLook, ...JSON.parse(localStorage.getItem(APP_KEY) || "{}") };

function applyLook() {
  localStorage.setItem(APP_KEY, JSON.stringify(look));

  // system theme follows the os preference
  let theme = look.theme;
  if (theme === "system") {
    theme = window.matchMedia("(prefers-color-scheme: light)").matches ? "light" : "dark";
  }
  document.documentElement.dataset.theme = theme;
  document.documentElement.dataset.style = look.style;
  document.documentElement.style.setProperty("--accent", look.accent);
  document.documentElement.style.setProperty("--radius", look.radius + "px");
  document.documentElement.style.setProperty("--btn-radius", look.btnRadius + "px");

  // theme quick button shows the opposite icon
  $("#themeBtn").innerHTML = icon(theme === "dark" ? "sun" : "moon");

  $$("#themeOptions .opt").forEach((b) => b.classList.toggle("active", b.dataset.theme === look.theme));
  $$("#styleOptions .opt").forEach((b) => b.classList.toggle("active", b.dataset.style === look.style));
  $$("#accentOptions .swatch").forEach((b) => b.classList.toggle("active", b.dataset.accent?.toLowerCase() === look.accent.toLowerCase()));
  $("#radiusRange").value = look.radius;
  $("#radiusVal").textContent = look.radius + "px";
  $("#btnRadiusRange").value = look.btnRadius;
  $("#btnRadiusVal").textContent = look.btnRadius + "px";
}

$$("#themeOptions .opt").forEach((b) => b.addEventListener("click", () => { look.theme = b.dataset.theme; applyLook(); }));
$$("#styleOptions .opt").forEach((b) => b.addEventListener("click", () => { look.style = b.dataset.style; applyLook(); }));
$$("#accentOptions .swatch").forEach((b) => b.addEventListener("click", () => {
  if (!b.dataset.accent) return;
  look.accent = b.dataset.accent; applyLook();
}));
$("#customAccent").addEventListener("input", (e) => { look.accent = e.target.value; applyLook(); });
$("#radiusRange").addEventListener("input", (e) => { look.radius = +e.target.value; applyLook(); });
$("#btnRadiusRange").addEventListener("input", (e) => { look.btnRadius = +e.target.value; applyLook(); });

// quick theme toggle in the navbar cycles dark -> light -> dark
$("#themeBtn").addEventListener("click", () => {
  look.theme = (document.documentElement.dataset.theme === "dark") ? "light" : "dark";
  applyLook();
});

// follow os changes while in system mode
window.matchMedia("(prefers-color-scheme: light)").addEventListener("change", () => {
  if (look.theme === "system") applyLook();
});

// ---------- boot ----------

applyLook();
$("#logoutBtn").innerHTML = icon("logout");

if (getToken()) {
  enterApp();
} else {
  showLogin();
}
