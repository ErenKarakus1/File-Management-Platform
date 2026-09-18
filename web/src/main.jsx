import React, { useEffect, useMemo, useRef, useState } from "react";
import { createRoot } from "react-dom/client";
import {
  CheckCircle2,
  Download,
  File,
  FileClock,
  FilePlus2,
  Loader2,
  LogOut,
  RefreshCw,
  ShieldAlert,
  Trash2,
  Upload,
  UserPlus,
} from "lucide-react";
import "./styles.css";

const API_BASE = "/api/v1";
const tokenKey = "fmp_access_token";

function App() {
  const [token, setToken] = useState(() => localStorage.getItem(tokenKey) || "");
  const [user, setUser] = useState(null);
  const [authMode, setAuthMode] = useState("login");
  const [authError, setAuthError] = useState("");
  const [files, setFiles] = useState([]);
  const [loadingFiles, setLoadingFiles] = useState(false);
  const [uploading, setUploading] = useState(false);
  const [notice, setNotice] = useState("");
  const fileInputRef = useRef(null);

  const api = useMemo(() => createApi(token), [token]);

  useEffect(() => {
    if (!token) {
      setUser(null);
      setFiles([]);
      return;
    }

    api.me()
      .then(setUser)
      .catch(() => logout());
  }, [token]);

  useEffect(() => {
    if (!user) return;
    loadFiles();
    const interval = window.setInterval(loadFiles, 5000);
    return () => window.clearInterval(interval);
  }, [user]);

  function saveToken(nextToken) {
    localStorage.setItem(tokenKey, nextToken);
    setToken(nextToken);
  }

  function logout() {
    localStorage.removeItem(tokenKey);
    setToken("");
    setUser(null);
    setFiles([]);
  }

  async function submitAuth(event) {
    event.preventDefault();
    setAuthError("");

    const formElement = event.currentTarget;
    const form = new FormData(formElement);
    const payload = {
      email: form.get("email"),
      password: form.get("password"),
    };
    if (authMode === "register") {
      payload.name = form.get("name");
    }

    try {
      const result = authMode === "register" ? await api.register(payload) : await api.login(payload);
      saveToken(result.access_token);
      setUser(result.user);
      formElement.reset();
    } catch (error) {
      setAuthError(error.message);
    }
  }

  async function loadFiles() {
    setLoadingFiles(true);
    try {
      const result = await api.listFiles();
      setFiles(result.files || []);
    } catch (error) {
      setNotice(error.message);
    } finally {
      setLoadingFiles(false);
    }
  }

  async function uploadFile(event) {
    const selected = event.target.files?.[0];
    if (!selected) return;

    setUploading(true);
    setNotice("");
    try {
      await api.uploadFile(selected);
      await loadFiles();
      setNotice("Upload queued for processing.");
    } catch (error) {
      setNotice(error.message);
    } finally {
      setUploading(false);
      event.target.value = "";
    }
  }

  async function deleteFile(fileID) {
    setNotice("");
    try {
      await api.deleteFile(fileID);
      await loadFiles();
      setNotice("Delete queued.");
    } catch (error) {
      setNotice(error.message);
    }
  }

  if (!token || !user) {
    return (
      <main className="auth-shell">
        <section className="auth-panel">
          <div className="brand-row">
            <div className="brand-mark">
              <FilePlus2 size={24} />
            </div>
            <div>
              <h1>File Management</h1>
              <p>Secure uploads, tracked processing, controlled downloads.</p>
            </div>
          </div>

          <div className="mode-tabs" role="tablist">
            <button className={authMode === "login" ? "active" : ""} onClick={() => setAuthMode("login")}>
              <ShieldAlert size={16} />
              Login
            </button>
            <button className={authMode === "register" ? "active" : ""} onClick={() => setAuthMode("register")}>
              <UserPlus size={16} />
              Register
            </button>
          </div>

          <form className="auth-form" onSubmit={submitAuth}>
            {authMode === "register" && (
              <label>
                Name
                <input name="name" type="text" minLength="3" maxLength="50" required />
              </label>
            )}
            <label>
              Email
              <input name="email" type="email" autoComplete="email" required />
            </label>
            <label>
              Password
              <input name="password" type="password" autoComplete={authMode === "login" ? "current-password" : "new-password"} minLength="8" required />
            </label>
            {authError && <p className="error-text">{authError}</p>}
            <button className="primary-button" type="submit">
              {authMode === "login" ? "Login" : "Create account"}
            </button>
          </form>
        </section>
      </main>
    );
  }

  return (
    <main className="app-shell">
      <header className="topbar">
        <div>
          <h1>Files</h1>
          <p>{user.name} · {user.email}</p>
        </div>
        <div className="topbar-actions">
          <button className="icon-button" title="Refresh files" onClick={loadFiles} disabled={loadingFiles}>
            <RefreshCw size={18} className={loadingFiles ? "spin" : ""} />
          </button>
          <button className="icon-button" title="Log out" onClick={logout}>
            <LogOut size={18} />
          </button>
        </div>
      </header>

      <section className="toolbar">
        <input ref={fileInputRef} type="file" onChange={uploadFile} hidden />
        <button className="primary-button" onClick={() => fileInputRef.current?.click()} disabled={uploading}>
          {uploading ? <Loader2 size={18} className="spin" /> : <Upload size={18} />}
          Upload file
        </button>
        {notice && <p className="notice">{notice}</p>}
      </section>

      <section className="file-table" aria-label="Files">
        <div className="file-row header">
          <span>Name</span>
          <span>Status</span>
          <span>Size</span>
          <span>Checksum</span>
          <span></span>
        </div>
        {files.length === 0 ? (
          <div className="empty-state">
            <FilePlus2 size={32} />
            <p>No files yet.</p>
          </div>
        ) : (
          files.map((file) => (
            <div className="file-row" key={file.id}>
              <div className="file-name">
                <File size={18} />
                <div>
                  <strong>{file.original_name}</strong>
                  <small>{file.content_type}</small>
                </div>
              </div>
              <StatusPill status={file.status} />
              <span>{formatBytes(file.size_bytes)}</span>
              <code title={file.checksum_sha256 || ""}>{file.checksum_sha256 ? shortChecksum(file.checksum_sha256) : "pending"}</code>
              <div className="row-actions">
                <a className={`icon-button ${file.status !== "ready" ? "disabled" : ""}`} title="Download file" href={file.status === "ready" ? `${API_BASE}/files/${file.id}/download` : undefined}>
                  <Download size={18} />
                </a>
                <button className="icon-button danger" title="Delete file" onClick={() => deleteFile(file.id)} disabled={file.status === "deleting"}>
                  <Trash2 size={18} />
                </button>
              </div>
            </div>
          ))
        )}
      </section>
    </main>
  );
}

function StatusPill({ status }) {
  const icon = status === "ready" ? <CheckCircle2 size={15} /> : status === "failed" ? <ShieldAlert size={15} /> : <FileClock size={15} />;
  return <span className={`status-pill ${status}`}>{icon}{status}</span>;
}

function createApi(token) {
  async function request(path, options = {}) {
    const headers = new Headers(options.headers || {});
    if (token) headers.set("Authorization", `Bearer ${token}`);
    if (options.body && !(options.body instanceof FormData)) {
      headers.set("Content-Type", "application/json");
    }

    const response = await fetch(`${API_BASE}${path}`, {
      ...options,
      headers,
    });

    if (response.status === 204) return null;
    const data = await response.json().catch(() => ({}));
    if (!response.ok) {
      throw new Error(data.error || "Request failed");
    }
    return data;
  }

  return {
    register: (payload) => request("/auth/register", { method: "POST", body: JSON.stringify(payload) }),
    login: (payload) => request("/auth/login", { method: "POST", body: JSON.stringify(payload) }),
    me: () => request("/auth/me"),
    listFiles: () => request("/files"),
    uploadFile: (file) => {
      const body = new FormData();
      body.append("file", file);
      return request("/files", { method: "POST", body });
    },
    deleteFile: (id) => request(`/files/${id}`, { method: "DELETE" }),
  };
}

function formatBytes(bytes) {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`;
}

function shortChecksum(checksum) {
  return `${checksum.slice(0, 10)}...${checksum.slice(-6)}`;
}

createRoot(document.getElementById("root")).render(<App />);
