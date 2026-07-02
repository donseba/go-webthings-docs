(() => {
    const escapeHTML = (value) => value
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;");

    const stringPattern = /("(?:\\.|[^"\\])*"|'(?:\\.|[^'\\])*'|`[\s\S]*?`)/g;
    const commentPattern = /(\/\/[^\n]*|\/\*[\s\S]*?\*\/|&lt;!--[\s\S]*?--&gt;)/g;
    const numberPattern = /\b(\d+(?:\.\d+)?)\b/g;
    const goKeywordPattern = /\b(break|case|chan|const|continue|default|defer|else|fallthrough|for|func|go|goto|if|import|interface|map|package|range|return|select|struct|switch|type|var|nil|true|false)\b/g;
    const templateKeywordPattern = /\b(define|template|block|if|else|end|range|with|printf|html|js|urlquery)\b/g;
    const jsKeywordPattern = /\b(await|async|break|case|catch|class|const|continue|default|else|export|for|from|function|if|import|let|new|null|return|switch|throw|try|undefined|var|while|true|false)\b/g;

    const inferLanguage = (text) => {
        if (text.includes("{{") || text.includes("&#123;&#123;") || text.includes("@model") || text.includes("@dot")) {
            return "template";
        }
        if (text.includes("<") && text.includes(">") && /<\/?[a-z][\s\S]*>/i.test(text)) {
            return "html";
        }
        if (/\b(func|package|type|struct|interface|defer|go)\b/.test(text) || text.includes(":=")) {
            return "go";
        }
        if (/\b(const|let|function|document|window|addEventListener)\b/.test(text)) {
            return "js";
        }
        if (/^\s*[{[]/.test(text)) {
            return "json";
        }
        return "plain";
    };

    const span = (className, value) => `<span class="${className}">${escapeHTML(value)}</span>`;
    const mark = (html, pattern, className) => html.replace(pattern, (match) => `<span class="${className}">${match}</span>`);

    const tokenKey = (index) => {
        let value = index;
        let key = "";
        do {
            key = String.fromCharCode(97 + (value % 26)) + key;
            value = Math.floor(value / 26) - 1;
        } while (value >= 0);
        return `\uE000${key}\uE001`;
    };

    const tokenIndex = (key) => {
        let value = 0;
        for (const ch of key) {
            value = value * 26 + (ch.charCodeAt(0) - 96);
        }
        return value - 1;
    };

    const stash = (html, pattern, className, tokens) => html.replace(pattern, (match) => {
        const key = tokenKey(tokens.length);
        tokens.push(`<span class="${className}">${match}</span>`);
        return key;
    });

    const restore = (html, tokens) => html.replace(/\uE000([a-z]+)\uE001/g, (_, key) => tokens[tokenIndex(key)]);

    const highlightType = (value) => {
        const lastDot = value.lastIndexOf(".");
        if (lastDot < 0) {
            return span("syntax-type", value);
        }
        return span("syntax-type-path", value.slice(0, lastDot + 1)) + span("syntax-type", value.slice(lastDot + 1));
    };

    const highlightAnnotationLine = (line) => {
        const match = line.match(/^(\s*)(@[a-zA-Z_][\w-]*)(?:\s+([^\s*]+))?(?:\s+([^\s*]+))?(.*)$/);
        if (!match) {
            return span("syntax-comment", line);
        }

        const [, indent, annotation, first = "", second = "", rest = ""] = match;
        let html = escapeHTML(indent) + span("syntax-annotation", annotation);
        if (annotation === "@dot") {
            if (first) {
                html += " " + highlightType(first);
            }
            if (second) {
                html += " " + span("syntax-comment", second);
            }
        } else {
            if (first) {
                html += " " + span("syntax-symbol", first);
            }
            if (second) {
                html += " " + highlightType(second);
            }
        }
        if (rest) {
            html += span("syntax-comment", rest);
        }
        return html;
    };

    const highlightGoDocComment = (action) => {
        const body = action.slice(4, -4);
        return span("syntax-template", "{{/*")
            + body.split("\n").map(highlightAnnotationLine).join("\n")
            + span("syntax-template", "*/}}");
    };

    const highlightTemplateAction = (action) => {
        if (action.startsWith("{{/*") && action.endsWith("*/}}")) {
            return highlightGoDocComment(action);
        }

        const tokens = [];
        let inner = escapeHTML(action.slice(2, -2));
        inner = stash(inner, stringPattern, "syntax-string", tokens);
        inner = mark(inner, templateKeywordPattern, "syntax-keyword");
        inner = mark(inner, /(\.[A-Za-z_]\w*)/g, "syntax-field");
        inner = mark(inner, /(\$[A-Za-z_]\w*)/g, "syntax-symbol");
        return span("syntax-template", "{{") + restore(inner, tokens) + span("syntax-template", "}}");
    };

    const highlightHTMLTag = (tag) => {
        const match = tag.match(/^(<\/?)([A-Za-z][\w:-]*)([\s\S]*?)(\/?>)$/);
        if (!match) {
            return escapeHTML(tag);
        }
        const [, open, name, attrs, close] = match;
        const tokens = [];
        let attrHTML = escapeHTML(attrs);
        attrHTML = stash(attrHTML, /("(?:&quot;|[^"])*"|'(?:&#39;|[^'])*')/g, "syntax-string", tokens);
        attrHTML = attrHTML.replace(/\b([A-Za-z_:][\w:.-]*)(=)/g, '<span class="syntax-attr">$1</span>$2');
        attrHTML = restore(attrHTML, tokens);
        return span("syntax-template", open) + span("syntax-tag", name) + attrHTML + span("syntax-template", close);
    };

    const highlightTemplate = (text) => {
        let html = "";
        let lastIndex = 0;
        const tokenPattern = /({{[\s\S]*?}}|<\/?[A-Za-z][^>\n]*>)/g;
        for (const match of text.matchAll(tokenPattern)) {
            html += escapeHTML(text.slice(lastIndex, match.index));
            const token = match[0];
            html += token.startsWith("{{") ? highlightTemplateAction(token) : highlightHTMLTag(token);
            lastIndex = match.index + token.length;
        }
        html += escapeHTML(text.slice(lastIndex));
        return html;
    };

    const highlight = (code) => {
        const text = code.textContent;
        const language = inferLanguage(text);

        if (language === "template" || language === "html") {
            code.innerHTML = highlightTemplate(text);
            code.dataset.highlighted = "true";
            code.dataset.language = language;
            return;
        }

        let html = escapeHTML(text);

        const tokens = [];
        html = stash(html, commentPattern, "syntax-comment", tokens);
        html = stash(html, stringPattern, "syntax-string", tokens);
        html = mark(html, numberPattern, "syntax-number");

        if (language === "go") {
            html = mark(html, goKeywordPattern, "syntax-keyword");
        } else if (language === "js") {
            html = mark(html, jsKeywordPattern, "syntax-keyword");
        }

        code.innerHTML = restore(html, tokens);
        code.dataset.highlighted = "true";
        code.dataset.language = language;
    };

    const highlightAll = (root = document) => {
        root.querySelectorAll("pre code:not([data-highlighted])").forEach(highlight);
    };

    document.addEventListener("DOMContentLoaded", () => highlightAll());
    document.addEventListener("htmx:afterSwap", (event) => highlightAll(event.target));
})();
