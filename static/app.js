(function () {
    "use strict";

    let users = [];
    let tags = [];
    let currentUserID = "";
    let editingCardID = null;

    const $ = (sel) => document.querySelector(sel);
    const $$ = (sel) => document.querySelectorAll(sel);

    // API helper
    async function api(method, path, body) {
        const opts = {
            method,
            headers: { "Content-Type": "application/json" },
        };
        if (body) opts.body = JSON.stringify(body);
        const res = await fetch("/api" + path, opts);
        const data = await res.json();
        if (data.error) {
            showError(data.error.message);
            throw data.error;
        }
        return data.data;
    }

    function showError(msg) {
        const toast = document.createElement("div");
        toast.className = "error-toast";
        toast.textContent = msg;
        document.body.appendChild(toast);
        setTimeout(() => toast.remove(), 3000);
    }

    // Init
    async function init() {
        users = await api("GET", "/users");
        tags = await api("GET", "/tags");
        populateUserSelectors();
        await loadBoard();
        pollNotifications();
        setInterval(pollNotifications, 30000);
    }

    function populateUserSelectors() {
        const selectors = ["#current-user", "#card-reporter", "#card-assignee"];
        for (const sel of selectors) {
            const el = $(sel);
            el.innerHTML = "";
            for (const u of users) {
                const opt = document.createElement("option");
                opt.value = u.id;
                opt.textContent = u.name;
                el.appendChild(opt);
            }
        }
        currentUserID = users.length > 0 ? users[0].id : "";
        $("#current-user").addEventListener("change", (e) => {
            currentUserID = e.target.value;
            pollNotifications();
        });
    }

    // Board
    async function loadBoard() {
        const cards = await api("GET", "/cards");
        renderBoard(cards);
    }

    function renderBoard(cards) {
        const columns = { todo: [], doing: [], done: [] };
        for (const c of cards) {
            if (columns[c.column]) columns[c.column].push(c);
        }
        for (const [col, list] of Object.entries(columns)) {
            const container = $(`#list-${col}`);
            container.innerHTML = "";
            $(`#count-${col}`).textContent = list.length;
            for (const card of list) {
                container.appendChild(renderCard(card));
            }
        }
    }

    function renderCard(card) {
        const el = document.createElement("div");
        el.className = "card";
        el.dataset.id = card.id;

        let tagsHTML = "";
        if (card.tags && card.tags.length > 0) {
            tagsHTML = '<div class="tag-badges">' +
                card.tags.map((t) => `<span class="tag-badge" style="background:${t.color}">${t.name}</span>`).join("") +
                "</div>";
        }

        const assignee = card.assignee || {};
        const initial = (assignee.name || "?")[0].toUpperCase();

        let subtaskHTML = "";
        if (card.subtask_total > 0) {
            subtaskHTML = `<span class="subtask-indicator">\u2611 ${card.subtask_completed}/${card.subtask_total}</span>`;
        }

        el.innerHTML = `
            ${tagsHTML}
            <div class="card-title">${escapeHTML(card.title)}</div>
            <div class="card-meta">
                <div class="card-meta-left">
                    <span class="priority-badge priority-${card.priority}">P${card.priority}</span>
                    ${subtaskHTML}
                </div>
                <span class="avatar" style="background:${assignee.avatar_color || '#666'}" title="${escapeHTML(assignee.name || '')}">${initial}</span>
            </div>
        `;

        el.addEventListener("click", () => openCardDetail(card.id));

        // Drag
        el.draggable = true;
        el.addEventListener("dragstart", (e) => {
            e.dataTransfer.setData("text/plain", card.id);
        });

        return el;
    }

    // Drag and drop
    for (const col of $$(".card-list")) {
        col.addEventListener("dragover", (e) => e.preventDefault());
        col.addEventListener("drop", async (e) => {
            e.preventDefault();
            const cardID = e.dataTransfer.getData("text/plain");
            const newCol = col.id.replace("list-", "");
            await api("PATCH", `/cards/${cardID}/move`, {
                column: newCol,
                user_id: currentUserID,
            });
            await loadBoard();
        });
    }

    // Modal
    function openModal() {
        $("#modal-overlay").style.display = "flex";
    }

    function closeModal() {
        $("#modal-overlay").style.display = "none";
        editingCardID = null;
    }

    $("#modal-close").addEventListener("click", closeModal);
    $("#modal-overlay").addEventListener("click", (e) => {
        if (e.target === $("#modal-overlay")) closeModal();
    });

    // New card
    $("#new-card-btn").addEventListener("click", () => {
        editingCardID = null;
        $("#modal-title").textContent = "New Card";
        $("#card-id").value = "";
        $("#card-title").value = "";
        $("#card-description").value = "";
        $("#card-priority").value = "3";
        $("#card-column").value = "todo";
        $("#card-reporter").value = currentUserID;
        $("#card-assignee").value = currentUserID;
        $("#card-detail-section").style.display = "none";
        $("#delete-card-btn").style.display = "none";
        openModal();
    });

    // Open card detail
    async function openCardDetail(cardID) {
        editingCardID = cardID;
        const card = await api("GET", `/cards/${cardID}`);
        $("#modal-title").textContent = "Edit Card";
        $("#card-id").value = card.id;
        $("#card-title").value = card.title;
        $("#card-description").value = card.description;
        $("#card-priority").value = card.priority;
        $("#card-column").value = card.column;
        $("#card-reporter").value = card.reporter_id;
        $("#card-assignee").value = card.assignee_id;
        $("#card-detail-section").style.display = "block";
        $("#delete-card-btn").style.display = "inline-block";
        renderTags(card.tags || []);
        renderSubtasks(card.subtasks || []);
        renderComments(card.comments || []);
        openModal();
    }

    // Save card
    $("#save-card-btn").addEventListener("click", async () => {
        const data = {
            title: $("#card-title").value,
            description: $("#card-description").value,
            priority: parseInt($("#card-priority").value),
            column: $("#card-column").value,
            reporter_id: $("#card-reporter").value,
            assignee_id: $("#card-assignee").value,
            user_id: currentUserID,
        };

        if (editingCardID) {
            data.sort_order = 0;
            await api("PUT", `/cards/${editingCardID}`, data);
        } else {
            await api("POST", "/cards", data);
        }
        closeModal();
        await loadBoard();
    });

    // Delete card
    $("#delete-card-btn").addEventListener("click", async () => {
        if (editingCardID && confirm("Delete this card?")) {
            await api("DELETE", `/cards/${editingCardID}?user_id=${currentUserID}`);
            closeModal();
            await loadBoard();
        }
    });

    // Tags
    function renderTags(cardTags) {
        const container = $("#card-tags");
        container.innerHTML = "";
        for (const t of cardTags) {
            const chip = document.createElement("span");
            chip.className = "tag-chip";
            chip.style.background = t.color;
            chip.innerHTML = `${escapeHTML(t.name)} <span class="remove-tag" data-tag-id="${t.id}">&times;</span>`;
            container.appendChild(chip);
        }

        container.querySelectorAll(".remove-tag").forEach((btn) => {
            btn.addEventListener("click", async () => {
                await api("DELETE", `/cards/${editingCardID}/tags/${btn.dataset.tagId}?user_id=${currentUserID}`);
                await openCardDetail(editingCardID);
            });
        });

        const select = $("#tag-select");
        select.innerHTML = "";
        const attached = new Set(cardTags.map((t) => t.id));
        for (const t of tags) {
            if (!attached.has(t.id)) {
                const opt = document.createElement("option");
                opt.value = t.id;
                opt.textContent = t.name;
                select.appendChild(opt);
            }
        }
    }

    $("#add-tag-btn").addEventListener("click", async () => {
        const tagID = $("#tag-select").value;
        if (!tagID) return;
        await api("POST", `/cards/${editingCardID}/tags`, {
            tag_id: tagID,
            user_id: currentUserID,
        });
        await openCardDetail(editingCardID);
    });

    // Subtasks
    function renderSubtasks(subtasks) {
        const container = $("#subtask-list");
        container.innerHTML = "";
        const done = subtasks.filter((s) => s.completed).length;
        $("#subtask-progress").textContent = subtasks.length > 0 ? `(${done}/${subtasks.length})` : "";

        for (const st of subtasks) {
            const item = document.createElement("div");
            item.className = "subtask-item";
            item.innerHTML = `
                <input type="checkbox" ${st.completed ? "checked" : ""} data-id="${st.id}" data-title="${escapeAttr(st.title)}">
                <span class="subtask-title ${st.completed ? "completed" : ""}">${escapeHTML(st.title)}</span>
                <span class="delete-subtask" data-id="${st.id}">&times;</span>
            `;
            container.appendChild(item);
        }

        container.querySelectorAll("input[type=checkbox]").forEach((cb) => {
            cb.addEventListener("change", async () => {
                await api("PUT", `/cards/${editingCardID}/subtasks/${cb.dataset.id}`, {
                    title: cb.dataset.title,
                    completed: cb.checked,
                    user_id: currentUserID,
                });
                await openCardDetail(editingCardID);
            });
        });

        container.querySelectorAll(".delete-subtask").forEach((btn) => {
            btn.addEventListener("click", async () => {
                await api("DELETE", `/cards/${editingCardID}/subtasks/${btn.dataset.id}?user_id=${currentUserID}`);
                await openCardDetail(editingCardID);
            });
        });
    }

    $("#add-subtask-btn").addEventListener("click", async () => {
        const input = $("#new-subtask-title");
        const title = input.value.trim();
        if (!title) return;
        await api("POST", `/cards/${editingCardID}/subtasks`, {
            title,
            user_id: currentUserID,
        });
        input.value = "";
        await openCardDetail(editingCardID);
    });

    // Comments
    function renderComments(comments) {
        const container = $("#comment-list");
        container.innerHTML = "";
        for (const c of comments) {
            const user = c.user || {};
            const item = document.createElement("div");
            item.className = "comment-item";
            item.innerHTML = `
                <div class="comment-header">
                    <span class="avatar" style="background:${user.avatar_color || '#666'};width:20px;height:20px;font-size:10px">${(user.name || "?")[0]}</span>
                    <span class="comment-author">${escapeHTML(user.name || "Unknown")}</span>
                    <span class="comment-time">${formatTime(c.created_at)}</span>
                    <button class="comment-delete" data-id="${c.id}">&times;</button>
                </div>
                <div class="comment-content">${formatMentions(escapeHTML(c.content))}</div>
            `;
            container.appendChild(item);
        }

        container.querySelectorAll(".comment-delete").forEach((btn) => {
            btn.addEventListener("click", async () => {
                await api("DELETE", `/cards/${editingCardID}/comments/${btn.dataset.id}?user_id=${currentUserID}`);
                await openCardDetail(editingCardID);
            });
        });
    }

    function formatMentions(text) {
        return text.replace(/@(\w+)/g, '<span class="mention">@$1</span>');
    }

    // Comment @mention autocomplete
    const commentInput = $("#new-comment");
    const mentionDropdown = $("#mention-dropdown");

    commentInput.addEventListener("input", () => {
        const val = commentInput.value;
        const cursor = commentInput.selectionStart;
        const before = val.substring(0, cursor);
        const match = before.match(/@(\w*)$/);

        if (match) {
            const query = match[1].toLowerCase();
            const filtered = users.filter((u) =>
                u.name.toLowerCase().startsWith(query)
            );
            if (filtered.length > 0) {
                mentionDropdown.innerHTML = "";
                for (const u of filtered) {
                    const opt = document.createElement("div");
                    opt.className = "mention-option";
                    opt.innerHTML = `<span class="avatar" style="background:${u.avatar_color};width:20px;height:20px;font-size:10px">${u.name[0]}</span> ${escapeHTML(u.name)}`;
                    opt.addEventListener("mousedown", (e) => {
                        e.preventDefault();
                        const prefix = before.substring(0, before.length - match[0].length);
                        const suffix = val.substring(cursor);
                        commentInput.value = prefix + "@" + u.name + " " + suffix;
                        mentionDropdown.style.display = "none";
                    });
                    mentionDropdown.appendChild(opt);
                }
                mentionDropdown.style.display = "block";
                return;
            }
        }
        mentionDropdown.style.display = "none";
    });

    commentInput.addEventListener("blur", () => {
        setTimeout(() => { mentionDropdown.style.display = "none"; }, 200);
    });

    $("#add-comment-btn").addEventListener("click", async () => {
        const content = commentInput.value.trim();
        if (!content) return;
        await api("POST", `/cards/${editingCardID}/comments`, {
            content,
            user_id: currentUserID,
        });
        commentInput.value = "";
        await openCardDetail(editingCardID);
    });

    // Notifications
    async function pollNotifications() {
        if (!currentUserID) return;
        const notifs = await api("GET", `/users/${currentUserID}/notifications?unread=true`);
        const badge = $("#notification-badge");
        if (notifs.length > 0) {
            badge.textContent = notifs.length;
            badge.style.display = "inline";
        } else {
            badge.style.display = "none";
        }
    }

    $("#notification-bell").addEventListener("click", async (e) => {
        e.stopPropagation();
        const dropdown = $("#notification-dropdown");
        const isOpen = dropdown.classList.contains("open");
        if (isOpen) {
            dropdown.classList.remove("open");
            return;
        }

        const notifs = await api("GET", `/users/${currentUserID}/notifications`);
        const list = $("#notification-list");
        list.innerHTML = "";
        if (notifs.length === 0) {
            list.innerHTML = '<div class="notif-item">No notifications</div>';
        } else {
            for (const n of notifs) {
                const item = document.createElement("div");
                item.className = "notif-item" + (n.read ? "" : " unread");
                item.innerHTML = `
                    <div>${escapeHTML(n.message)}</div>
                    <div class="notif-time">${formatTime(n.created_at)}</div>
                `;
                if (!n.read) {
                    item.style.cursor = "pointer";
                    item.addEventListener("click", async () => {
                        await api("PATCH", `/users/${currentUserID}/notifications/${n.id}/read`);
                        pollNotifications();
                        item.classList.remove("unread");
                    });
                }
                list.appendChild(item);
            }
        }
        dropdown.classList.add("open");
    });

    document.addEventListener("click", () => {
        $("#notification-dropdown").classList.remove("open");
    });

    $("#mark-all-read").addEventListener("click", async (e) => {
        e.stopPropagation();
        await api("PATCH", `/users/${currentUserID}/notifications/read-all`);
        pollNotifications();
        $$("#notification-list .notif-item").forEach((el) => el.classList.remove("unread"));
    });

    // Helpers
    function escapeHTML(str) {
        const div = document.createElement("div");
        div.textContent = str || "";
        return div.innerHTML;
    }

    function escapeAttr(str) {
        return (str || "").replace(/"/g, "&quot;").replace(/'/g, "&#39;");
    }

    function formatTime(iso) {
        if (!iso) return "";
        const d = new Date(iso);
        const now = new Date();
        const diff = (now - d) / 1000;
        if (diff < 60) return "just now";
        if (diff < 3600) return Math.floor(diff / 60) + "m ago";
        if (diff < 86400) return Math.floor(diff / 3600) + "h ago";
        return d.toLocaleDateString();
    }

    init();
})();
