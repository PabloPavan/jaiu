(() => {
  const sseKey = "__jaiuEventSource";
  let eventSource = null;
  if (window[sseKey] && typeof window[sseKey].close === "function") {
    window[sseKey].close();
  }
  window[sseKey] = null;
  function photoUpload(el) {
    const dataset = (el && el.dataset) || {};
    const initialObjectKey = dataset.photoObjectKey || "";
    const initialUrl = dataset.photoUrl || "";

    return {
      photoObjectKey: initialObjectKey,
      previewUrl: initialUrl,
      error: "",
      async handleFile(event) {
        const file = event && event.target && event.target.files && event.target.files[0];
        if (!file) {
          return;
        }
        this.error = "";

        if (this.previewUrl && this.previewUrl.startsWith("blob:")) {
          URL.revokeObjectURL(this.previewUrl);
        }
        this.previewUrl = URL.createObjectURL(file);
      },
      clear() {
        if (this.previewUrl && this.previewUrl.startsWith("blob:")) {
          URL.revokeObjectURL(this.previewUrl);
        }
        this.photoObjectKey = "";
        this.previewUrl = "";
        this.error = "";
        if (this.$refs && this.$refs.photoInput) {
          this.$refs.photoInput.value = "";
        }
      },
    };
  }

  function buildRefreshUrl(el) {
    const url = el && el.dataset ? el.dataset.sseUrl : "";
    if (!url) {
      return "";
    }
    if (url.includes("?")) {
      return url;
    }
    return url + window.location.search;
  }

  function getOriginId() {
    const storageKey = "jaiu-origin-id";
    let originId = window.sessionStorage ? window.sessionStorage.getItem(storageKey) : "";
    if (!originId) {
      if (window.crypto && typeof window.crypto.randomUUID === "function") {
        originId = window.crypto.randomUUID();
      } else {
        originId = `oid-${Date.now()}-${Math.random().toString(16).slice(2)}`;
      }
      if (window.sessionStorage) {
        window.sessionStorage.setItem(storageKey, originId);
      }
    }
    return originId;
  }

  function ensureOriginInput(form, originId) {
    if (!form || !originId) {
      return;
    }
    const existing = form.querySelector('input[name="origin_id"]');
    if (existing) {
      existing.value = originId;
      return;
    }
    const input = document.createElement("input");
    input.type = "hidden";
    input.name = "origin_id";
    input.value = originId;
    form.appendChild(input);
  }

  function withOriginQuery(url, originId) {
    if (!url || !originId) {
      return url;
    }
    let parsed = null;
    try {
      parsed = new URL(url, window.location.href);
    } catch (err) {
      return url;
    }
    if (parsed.searchParams.has("origin_id")) {
      parsed.searchParams.set("origin_id", originId);
      return parsed.toString();
    }
    parsed.searchParams.append("origin_id", originId);
    return parsed.toString();
  }

  function isHTMXForm(form) {
    if (!form) {
      return false;
    }
    const attrs = ["hx-post", "hx-put", "hx-patch", "hx-delete"];
    return attrs.some((attr) => form.hasAttribute(attr));
  }

  function setupOriginId(originId) {
    document.addEventListener("submit", (event) => {
      const form = event.target;
      const method = form && form.method ? form.method.toLowerCase() : "get";
      if (!form || method === "get") {
        return;
      }
      if (!isHTMXForm(form)) {
        form.action = withOriginQuery(form.action || "", originId);
      }
      ensureOriginInput(form, originId);
    });

    document.addEventListener("htmx:configRequest", (event) => {
      if (!event.detail || !event.detail.headers) {
        return;
      }
      event.detail.headers["X-Origin-ID"] = originId;
    });
  }

  function refreshElement(el) {
    const url = buildRefreshUrl(el);
    if (!url) {
      return;
    }
    if (window.htmx && typeof window.htmx.ajax === "function") {
      window.htmx.ajax("GET", url, { target: el, swap: "outerHTML" });
      return;
    }
    window.location.reload();
  }

  function setupSSE(originId) {
    if (!window.EventSource) {
      return;
    }

    const topics = new Set();
    document.querySelectorAll("[data-sse-topic]").forEach((el) => {
      const topic = el.dataset ? el.dataset.sseTopic : "";
      if (topic) {
        topics.add(topic);
      }
    });

    if (topics.size === 0) {
      return;
    }

    if (eventSource) {
      return;
    }
    eventSource = new EventSource("/events");
    window[sseKey] = eventSource;
    topics.forEach((topic) => {
      eventSource.addEventListener(`app.${topic}.changed`, (event) => {
        let data = null;
        try {
          data = JSON.parse(event.data || "{}");
        } catch (err) {
          data = null;
        }
        if (data && data.origin_id && data.origin_id === originId) {
          return;
        }
        document.querySelectorAll(`[data-sse-topic="${topic}"]`).forEach((el) => {
          refreshElement(el);
        });
      });
    });
  }

  window.photoUpload = photoUpload;
  const originId = getOriginId();
  setupOriginId(originId);
  setupSSE(originId);
  function closeEventSource() {
    if (!eventSource) {
      return;
    }
    eventSource.close();
    eventSource = null;
    if (window[sseKey]) {
      window[sseKey] = null;
    }
  }
  function shouldCloseForLink(link) {
    if (!link) {
      return false;
    }
    const target = link.getAttribute("target");
    if (target && target.toLowerCase() !== "_self") {
      return false;
    }
    if (link.hasAttribute("download")) {
      return false;
    }
    const href = link.getAttribute("href") || "";
    if (!href || href.startsWith("#")) {
      return false;
    }
    let url = null;
    try {
      url = new URL(href, window.location.href);
    } catch (err) {
      return false;
    }
    if (url.origin !== window.location.origin) {
      return false;
    }
    if (url.pathname === window.location.pathname && url.search === window.location.search && url.hash) {
      return false;
    }
    return true;
  }
  document.addEventListener("click", (event) => {
    const link = event.target && event.target.closest ? event.target.closest("a") : null;
    if (shouldCloseForLink(link)) {
      closeEventSource();
    }
  });
  document.addEventListener("visibilitychange", () => {
    if (document.visibilityState === "hidden") {
      closeEventSource();
      return;
    }
    if (document.visibilityState === "visible") {
      setupSSE(originId);
    }
  });
  window.addEventListener("pagehide", closeEventSource);
  window.addEventListener("beforeunload", closeEventSource);
  window.addEventListener("pageshow", () => {
    if (!eventSource) {
      setupSSE(originId);
    }
  });
})();
