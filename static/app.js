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
        if (users.length === 0) {
            showCreateUserPrompt();
            return;
        }
        populateUserSelectors();
        await loadBoard();
        pollNotifications();
        pollChatUnread();
        setInterval(pollNotifications, 30000);
        setInterval(pollChatUnread, 15000);
    }

    function showCreateUserPrompt() {
        const name = prompt("No users found. Enter your name to get started:");
        if (!name || !name.trim()) {
            showError("A user is required to use the board.");
            setTimeout(showCreateUserPrompt, 500);
            return;
        }
        const colors = ["#EF4444", "#3B82F6", "#10B981", "#F59E0B", "#8B5CF6", "#EC4899", "#06B6D4"];
        const color = colors[Math.floor(Math.random() * colors.length)];
        api("POST", "/users", { name: name.trim(), avatar_color: color }).then((user) => {
            users = [user];
            populateUserSelectors();
            loadBoard();
            pollNotifications();
            pollChatUnread();
            setInterval(pollNotifications, 30000);
            setInterval(pollChatUnread, 15000);
        });
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
            pollChatUnread();
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

        let childHTML = "";
        if (card.child_total > 0) {
            childHTML = `<span class="subtask-indicator">\u2611 ${card.child_completed}/${card.child_total}</span>`;
        }

        el.innerHTML = `
            ${tagsHTML}
            <div class="card-title">${escapeHTML(card.title)}</div>
            <div class="card-meta">
                <div class="card-meta-left">
                    <span class="priority-badge priority-${card.priority}">P${card.priority}</span>
                    ${childHTML}
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
        renderChildren(card.children || [], card.child_total, card.child_completed);
        loadFiles(card.id);
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

    // Children
    function renderChildren(children, total, completed) {
        const container = $("#children-list");
        container.innerHTML = "";
        $("#children-progress").textContent = total > 0 ? `(${completed}/${total})` : "";

        for (const child of children) {
            const assignee = child.assignee || {};
            const initial = (assignee.name || "?")[0].toUpperCase();
            const item = document.createElement("div");
            item.className = "subtask-item child-card-item";
            item.style.cursor = "pointer";
            const colLabel = child.column === "done" ? "\u2705" : child.column === "doing" ? "\uD83D\uDD35" : "\u26AA";
            item.innerHTML = `
                <span>${colLabel}</span>
                <span class="subtask-title ${child.column === "done" ? "completed" : ""}">${escapeHTML(child.title)}</span>
                <span class="avatar" style="background:${assignee.avatar_color || '#666'};width:20px;height:20px;font-size:10px" title="${escapeHTML(assignee.name || '')}">${initial}</span>
            `;
            item.addEventListener("click", () => openCardDetail(child.id));
            container.appendChild(item);
        }
    }

    $("#add-child-btn").addEventListener("click", async () => {
        const title = prompt("Child card title:");
        if (!title || !title.trim()) return;
        await api("POST", "/cards", {
            title: title.trim(),
            description: "",
            priority: 3,
            column: "todo",
            reporter_id: currentUserID,
            assignee_id: currentUserID,
            parent_id: editingCardID,
            user_id: currentUserID,
        });
        await openCardDetail(editingCardID);
        await loadBoard();
    });

    // Files
    async function loadFiles(cardID) {
        const res = await fetch("/api/cards/" + cardID + "/files");
        const data = await res.json();
        const files = data.data || [];
        renderFiles(files);
    }

    function renderFiles(files) {
        const container = $("#file-list");
        container.innerHTML = "";
        $("#file-count").textContent = files.length > 0 ? `(${files.length})` : "";

        for (const f of files) {
            const item = document.createElement("div");
            item.className = "file-item";
            const sizeStr = f.size < 1024 ? f.size + " B"
                : f.size < 1048576 ? (f.size / 1024).toFixed(1) + " KB"
                : (f.size / 1048576).toFixed(1) + " MB";
            item.innerHTML = `
                <a href="${f.raw_url}" target="_blank" class="file-link">${escapeHTML(f.filename)}</a>
                <span class="file-size">${sizeStr}</span>
                <span class="delete-file" data-id="${f.id}">&times;</span>
            `;
            container.appendChild(item);
        }

        container.querySelectorAll(".delete-file").forEach((btn) => {
            btn.addEventListener("click", async () => {
                await fetch("/api/files/" + btn.dataset.id + "?user_id=" + currentUserID, { method: "DELETE" });
                await loadFiles(editingCardID);
            });
        });
    }

    $("#upload-file-btn").addEventListener("click", () => {
        $("#file-input").click();
    });

    $("#file-input").addEventListener("change", async () => {
        const input = $("#file-input");
        if (!input.files.length) return;
        const formData = new FormData();
        formData.append("file", input.files[0]);
        formData.append("user_id", currentUserID);
        const res = await fetch("/api/cards/" + editingCardID + "/files", {
            method: "POST",
            body: formData,
        });
        const data = await res.json();
        if (data.error) {
            showError(data.error.message);
        }
        input.value = "";
        await loadFiles(editingCardID);
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
        return text.replace(/@([\w][\w-]*(?:\s[\w][\w-]*)?)/g, '<span class="mention">@$1</span>');
    }

    // Comment @mention autocomplete
    const commentInput = $("#new-comment");
    const mentionDropdown = $("#mention-dropdown");

    commentInput.addEventListener("input", () => {
        const val = commentInput.value;
        const cursor = commentInput.selectionStart;
        const before = val.substring(0, cursor);
        const match = before.match(/@([\w-]*)$/);

        if (match) {
            const query = normalizeMention(match[1]);
            const filtered = users.filter((u) =>
                normalizeMention(u.name).startsWith(query)
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

    // Chat
    let chatOpen = false;

    $("#chat-toggle").addEventListener("click", async (e) => {
        e.stopPropagation();
        chatOpen = !chatOpen;
        const panel = $("#chat-panel");
        if (chatOpen) {
            panel.classList.add("open");
            await loadChatMessages();
            await markChatRead();
            scrollChatBottom();
        } else {
            panel.classList.remove("open");
        }
    });

    $("#chat-close").addEventListener("click", () => {
        chatOpen = false;
        $("#chat-panel").classList.remove("open");
    });

    async function loadChatMessages() {
        const messages = await api("GET", "/messages?limit=50");
        const container = $("#chat-messages");
        container.innerHTML = "";
        const ordered = messages.reverse();
        for (const m of ordered) {
            container.appendChild(renderChatMessage(m));
        }
    }

    function renderChatMessage(m) {
        const user = m.user || {};
        const initial = (user.name || "?")[0].toUpperCase();
        const el = document.createElement("div");
        el.className = "chat-msg";
        el.innerHTML = `
            <span class="avatar" style="background:${user.avatar_color || '#666'};width:28px;height:28px;font-size:12px;flex-shrink:0">${initial}</span>
            <div class="chat-msg-body">
                <div class="chat-msg-header">
                    <span class="chat-msg-author">${escapeHTML(user.name || "Unknown")}</span>
                    <span class="chat-msg-time">${formatTime(m.created_at)}</span>
                </div>
                <div class="chat-msg-content">${formatMentions(escapeHTML(m.content))}</div>
            </div>
        `;
        return el;
    }

    function scrollChatBottom() {
        const container = $("#chat-messages");
        container.scrollTop = container.scrollHeight;
    }

    async function markChatRead() {
        await api("PATCH", "/messages/mark-read", { user_id: currentUserID });
        pollChatUnread();
    }

    async function pollChatUnread() {
        if (!currentUserID) return;
        const res = await api("GET", `/messages/unread-count?user_id=${currentUserID}`);
        const badge = $("#chat-badge");
        if (res.unread_count > 0) {
            badge.textContent = res.unread_count;
            badge.style.display = "inline";
        } else {
            badge.style.display = "none";
        }
    }

    // Chat @mention autocomplete
    const chatInput = $("#chat-input-text");
    const chatMentionDropdown = $("#chat-mention-dropdown");

    chatInput.addEventListener("input", () => {
        const val = chatInput.value;
        const cursor = chatInput.selectionStart;
        const before = val.substring(0, cursor);
        const match = before.match(/@([\w-]*)$/);

        if (match) {
            const query = normalizeMention(match[1]);
            const filtered = users.filter((u) =>
                normalizeMention(u.name).startsWith(query)
            );
            if (filtered.length > 0) {
                chatMentionDropdown.innerHTML = "";
                for (const u of filtered) {
                    const opt = document.createElement("div");
                    opt.className = "mention-option";
                    opt.innerHTML = `<span class="avatar" style="background:${u.avatar_color};width:20px;height:20px;font-size:10px">${u.name[0]}</span> ${escapeHTML(u.name)}`;
                    opt.addEventListener("mousedown", (e) => {
                        e.preventDefault();
                        const prefix = before.substring(0, before.length - match[0].length);
                        const suffix = val.substring(cursor);
                        chatInput.value = prefix + "@" + u.name + " " + suffix;
                        chatMentionDropdown.style.display = "none";
                    });
                    chatMentionDropdown.appendChild(opt);
                }
                chatMentionDropdown.style.display = "block";
                return;
            }
        }
        chatMentionDropdown.style.display = "none";
    });

    chatInput.addEventListener("blur", () => {
        setTimeout(() => { chatMentionDropdown.style.display = "none"; }, 200);
    });

    $("#chat-send-btn").addEventListener("click", async () => {
        const content = chatInput.value.trim();
        if (!content) return;
        await api("POST", "/messages", {
            content,
            user_id: currentUserID,
        });
        chatInput.value = "";
        await loadChatMessages();
        scrollChatBottom();
    });

    chatInput.addEventListener("keydown", (e) => {
        if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            $("#chat-send-btn").click();
        }
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
    function normalizeMention(str) {
        return (str || "").toLowerCase().replace(/[-_ ]/g, "");
    }

    function escapeHTML(str) {
        const div = document.createElement("div");
        div.textContent = str || "";
        return div.innerHTML;
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
