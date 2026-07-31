// Theme toggle, custom select, and clipboard. Delegated from document so
// htmx-swapped and hx-boost-replaced markup keeps working without rebinding.
// The flag matches the select snippet's guard, so a page with both binds once.
window.__qompntSelect = true;
window.__qompntSearch = true;
// Variant options on a detail page. A control names a target inside
// [data-preview] and the classes to swap, so the page never holds two copies of
// the same component. The code block below the preview is printed from that same
// DOM, so it always shows what is on screen.
const words = (s) => (s || "").split(/\s+/).filter(Boolean);

// Print the preview as source: drop the demo-only data-x-* hooks and anything
// currently hidden, then break tags onto their own lines.
function previewSource(preview) {
    const copy = preview.cloneNode(true);
    copy.querySelectorAll("[hidden]").forEach((n) => n.remove());
    copy.querySelectorAll("*").forEach((n) => {
        [...n.attributes].forEach((a) => {
            // data-x-* are demo hooks; an empty class is what toggling every class off
            // leaves behind.
            if (
                a.name.startsWith("data-x-") ||
                (a.name === "class" && !a.value.trim())
            ) {
                n.removeAttribute(a.name);
            }
        });
    });

    const html = copy.innerHTML
        .replace(/\s+/g, " ")
        .replace(/> </g, ">\n<")
        .trim();
    let depth = 0;
    return html
        .split("\n")
        .map((line) => {
            if (line.startsWith("</")) depth--;
            const out = "  ".repeat(Math.max(depth, 0)) + line;
            const opens = /^<[a-zA-Z]/.test(line);
            const selfCloses =
                /\/>$/.test(line) ||
                /^<(input|img|br|hr|path|circle|meta|link)\b/.test(line);
            const closesOnSameLine = /<\/[a-zA-Z-]+>$/.test(line);
            if (opens && !selfCloses && !closesOnSameLine) depth++;
            return out;
        })
        .join("\n");
}

// Re-apply every option. A component that swaps itself over htmx (tabs, the
// validated field) comes back as fresh server markup with none of the option
// classes on it, so the panel has to write them again.
//
// This is watched by a MutationObserver, and observers fire asynchronously - a
// synchronous "am I already running" flag does not hold across that gap. So the
// observer is disconnected for the duration instead. Without it, a text option
// writing textContent triggers the observer, which re-applies, which writes
// again: an infinite loop that pins a core and eats memory until the tab dies.
// One page can hold more than one component - Button renders Button Group below
// it - so every lookup is scoped to the [data-block] a control lives in. Without
// that, the second panel's controls would write onto the first preview and the
// code block under one would print the other.
let previewObservers = [];

const blocksOf = () => [...document.querySelectorAll("[data-block]")];

function applyAllOptions() {
    previewObservers.forEach((o) => o.disconnect());
    document
        .querySelectorAll("[data-options] [data-opt-target]")
        .forEach((c) => {
            c.dispatchEvent(new Event("input", { bubbles: true }));
        });
    observePreview();
}

function observePreview() {
    previewObservers.forEach((o) => {
        const preview = o.__target;
        if (preview?.isConnected)
            o.observe(preview, { childList: true, subtree: true });
    });
}

// Only the Tailwind pane tracks the options - it is the one showing the live
// markup. The HTML and CSS panes are the server's rendering of the same
// component and do not change with a toggle.
function syncCode() {
    blocksOf().forEach((block) => {
        const preview = block.querySelector("[data-preview]");
        const code = block.querySelector("[data-code] code");
        if (preview && code) code.textContent = previewSource(preview);
    });
}

document.addEventListener("input", (e) => {
    const ctl = e.target.closest("[data-options] [data-opt-target]");
    if (!ctl) return;
    // The preview this control belongs to, not the first one on the page.
    const preview = ctl
        .closest("[data-block]")
        ?.querySelector("[data-preview]");
    if (!preview) return;

    const apply = (c) => {
        preview.querySelectorAll(c.dataset.optTarget).forEach((t) => {
            // A slider picks one class set out of data-opt-values, pipe separated.
            if (c.type === "range") {
                const sets = (c.dataset.optValues || "").split("|");
                sets.forEach((s) => t.classList.remove(...words(s)));
                t.classList.add(...words(sets[Number(c.value)] || ""));
                return;
            }
            if (c.type !== "checkbox" && c.type !== "radio") {
                const text = c.value.trim() || c.placeholder;
                // A text option writes words into the preview. Where those words live
                // depends on what it is pointed at: a field has no text content to set -
                // assigning textContent to an <input> silently does nothing - so the
                // words go to its placeholder instead. Combobox's Placeholder option was
                // dead for exactly this reason.
                if (t.tagName === "INPUT" || t.tagName === "TEXTAREA")
                    t.placeholder = text;
                else t.textContent = text;
                return;
            }
            // toggleAttribute, not .hidden: SVG elements do not reflect the property.
            if (c.hasAttribute("data-opt-hide"))
                t.toggleAttribute("hidden", !c.checked);
            // Some options are not a look, they are a behaviour: Slider's stepped mode
            // is the `step` attribute and nothing else. Classes cannot express that,
            // so an option may name an attribute and the value for each state.
            if (c.dataset.optAttr) {
                // Presence decides, not the value. `disabled` is a boolean attribute
                // whose correct value is the empty string, so "" has to mean set it -
                // omitting data-opt-attr-off is how you say remove it.
                const key = c.checked ? "optAttrOn" : "optAttrOff";
                if (key in c.dataset)
                    t.setAttribute(c.dataset.optAttr, c.dataset[key]);
                else t.removeAttribute(c.dataset.optAttr);
            }
            const on = words(c.dataset.optOn);
            const off = words(c.dataset.optOff);
            t.classList.add(...(c.checked ? on : off));
            t.classList.remove(...(c.checked ? off : on));
        });
    };

    // A radio only fires for the one being turned on, so run the whole group -
    // otherwise the losing option's classes stay applied.
    if (ctl.type === "radio" && ctl.name) {
        ctl.closest("[data-options]")
            .querySelectorAll(
                `input[type="radio"][name="${ctl.name}"][data-opt-target]`,
            )
            .forEach(apply);
    } else {
        apply(ctl);
    }
    syncCode();
});

// Combobox: filter on type, pick with click or arrows, close on Escape or an
// outside click. Same handler the component ships with.
window.__qompntCombobox = true;
const cbOptions = (box) => [...box.querySelectorAll('[role="option"]')];

document.addEventListener("input", (e) => {
    const input = e.target.closest('[data-combobox] input[role="combobox"]');
    if (!input) return;
    const box = input.closest("[data-combobox]");
    const q = input.value.trim().toLowerCase();
    let shown = 0;
    cbOptions(box).forEach((o) => {
        const hit = o.textContent.trim().toLowerCase().includes(q);
        o.hidden = !hit;
        if (hit) shown++;
    });
    const empty = box.querySelector("[data-combobox-empty]");
    if (empty) empty.hidden = shown > 0;
    box.querySelector('[role="listbox"]').hidden = false;
    input.setAttribute("aria-expanded", "true");
});

document.addEventListener("keydown", (e) => {
    const box = e.target.closest("[data-combobox]");
    if (!box) return;
    if (e.key === "Escape") {
        box.querySelector('[role="listbox"]').hidden = true;
        return;
    }
    if (e.key !== "ArrowDown" && e.key !== "ArrowUp") return;
    e.preventDefault();
    const list = cbOptions(box).filter((o) => !o.hidden);
    const i = list.indexOf(document.activeElement);
    const next = e.key === "ArrowDown" ? i + 1 : i - 1;
    (list[next] || list[e.key === "ArrowDown" ? 0 : list.length - 1])?.focus();
});

// The design system is a stylesheet of tokens layered over the contract. Every
// component reads those tokens, so this is the entire switcher.
//
// Not every system has both schemes: some source docs never document a dark
// palette, and one is dark-first. Those declare what they have, and the toggle
// is disabled rather than pretending to switch something that does not exist.
function schemesFor(name) {
    const opt = document.querySelector(`[data-ds-option="${name || ""}"]`);
    return (opt?.dataset.dsSchemes || "light dark").split(" ");
}

// Rewriting the tokens should land in one repaint. Without this, every element
// on the page runs its own colour transition and the swap becomes a second-long
// cross-fade of the entire UI - which is why this used to be a blanket
// `transition: none` in tokens.css, at the cost of the components' own motion.
//
// Two frames, not one: the attribute has to survive the frame the new styles are
// computed in. Removing it in the same frame as the change lets the browser
// coalesce both into a single style recalculation, and the transition runs after
// all - the bug this is here to prevent.
function withoutTransitions(fn) {
    const root = document.documentElement;
    root.dataset.swapping = "";
    fn();
    requestAnimationFrame(() =>
        requestAnimationFrame(() => delete root.dataset.swapping),
    );
}

// The build stamp the page was served with, lifted off an asset the layout
// already stamped. A theme stylesheet loaded from here has to carry it too, or
// it is the one file on the site the browser has to revalidate every time.
const assetV =
    new URL(
        document.querySelector('link[href*="tokens.css"]').href,
    ).searchParams.get("v") || "";

function applyDesignSystem(name) {
    // removeAttribute rather than href='' inside: an empty href points at the
    // current page, which the browser dutifully downloads and parses as CSS.
    withoutTransitions(() => applyDesignSystemLink(name));
    try {
        if (name) localStorage.setItem("qompnt-ds", name);
        else localStorage.removeItem("qompnt-ds");
    } catch (_) {}
    enforceScheme(name);
}

// Move to a scheme this system actually has, and disable the toggle when there
// is only one. The user's own preference is remembered, so switching back to a
// system with both restores it.
function enforceScheme(name) {
    const schemes = schemesFor(name);
    const root = document.documentElement;
    let wanted = root.dataset.theme;
    try {
        wanted = localStorage.getItem("qompnt-theme") || wanted;
    } catch (_) {}

    root.dataset.theme = schemes.includes(wanted) ? wanted : schemes[0];

    document.querySelectorAll("[data-theme-toggle]").forEach((btn) => {
        const single = schemes.length === 1;
        btn.disabled = single;
        btn.setAttribute("aria-disabled", String(single));
        btn.classList.toggle("opacity-40", single);
        btn.setAttribute(
            "title",
            single ? `This system is ${schemes[0]}-only` : "Toggle dark mode",
        );
    });
}

// The page renders with the default selected; reflect the stored choice.
function syncDesignSystem() {
    const box = document.querySelector("[data-ds-switch]");
    if (!box) return;
    const current = document.documentElement.dataset.ds || "";
    const opt = box.querySelector(`[data-ds-option="${current}"]`);
    if (!opt) return;
    box.querySelectorAll('[role="option"]').forEach((o) =>
        o.setAttribute("aria-selected", String(o === opt)),
    );
    box.querySelector("[data-select-label]").textContent =
        opt.textContent.trim();
    enforceScheme(current);
}

// Motion. One attribute on <html>; tokens.css zeroes the duration tokens from
// there and clamps anything carrying its own timing.
//
// The switch ships unchecked and this is what checks it, rather than the server
// rendering the state: the answer depends on localStorage and on the visitor's
// prefers-reduced-motion, neither of which the server knows. It also has to run
// again after every hx-boost navigation, which hands back fresh markup.
// A range in the Customize panel is the one control whose setting you cannot
// read off the page - a slider at two-thirds tells you nothing about what it
// selected. The readout is written here rather than into twenty-five
// options.html files, and the label comes from data-opt-values when the range
// picks between named sets.
function syncOptionValues() {
    document
        .querySelectorAll('[data-options] input[type="range"]')
        .forEach((r) => {
            // The panel eats its own cooking: this is the shipped Slider, class and
            // all, so it is drawn by components/slider/styles.css rather than by a
            // second copy of those rules. --p is the fill percentage the component's
            // track gradient reads - the shipped markup sets it from an inline
            // oninput, and the panel has no markup of its own, so it is set here.
            r.classList.add("slider");
            const span = r.max - r.min || 1;
            r.style.setProperty("--p", ((r.value - r.min) / span) * 100);

            let out = r.parentElement.querySelector("[data-opt-readout]");
            if (!out) {
                out = document.createElement("span");
                out.dataset.optReadout = "";
                out.className =
                    "shrink-0 tabular-nums text-caption text-muted-foreground";
                r.after(out);
            }
            const sets = (r.dataset.optValues || "").split("|");
            // A class-set range is a list of named steps; anything else is a number.
            out.textContent = r.dataset.optValues
                ? String(sets[Number(r.value)] || "")
                      .split(" ")[0]
                      .replace(/^text-|^w-/, "")
                : r.value;
        });
}

document.addEventListener("input", (e) => {
    if (e.target.closest('[data-options] input[type="range"]'))
        syncOptionValues();
});

function syncMotion() {
    const on = document.documentElement.dataset.motion !== "off";
    document.querySelectorAll("[data-motion-toggle]").forEach((box) => {
        box.checked = on;
    });
}

// Stored explicitly as 'on' or 'off'. A visitor who turns motion on despite
// prefers-reduced-motion keeps it - the OS setting is the default here, not a
// veto - and tokens.css reads an explicit 'on' as exactly that override.
document.addEventListener("change", (e) => {
    const box = e.target.closest("[data-motion-toggle]");
    if (!box) return;
    const next = box.checked ? "on" : "off";
    document.documentElement.dataset.motion = next;
    try {
        localStorage.setItem("qompnt-motion", next);
    } catch (_) {}
});

// A toast stack is a queue, not a log: past four, drop the oldest. Without this
// a held-down button fills the viewport and none of them are readable.
const TOAST_MAX = 4;
function trimToasts() {
    const stack = document.getElementById("toasts");
    if (!stack) return;
    while (stack.children.length > TOAST_MAX) stack.lastElementChild.remove();
}
// A MutationObserver, not an htmx event: this holds however the toast arrives -
// htmx swap, a fetch, or app code appending to the stack.
const toastStack = document.getElementById("toasts");
if (toastStack)
    new MutationObserver(trimToasts).observe(toastStack, { childList: true });

// hx-boost replaces the body, so re-print on every settle as well as on load.
document.addEventListener("DOMContentLoaded", () => {
    syncCode();
    syncDesignSystem();
    syncMotion();
    syncOptionValues();
    // The preview can replace itself (htmx); when it does, the options go back on.
    // Only element insertions count: a text option setting textContent also
    // mutates the tree, and reacting to that is what makes it loop.
    previewObservers = blocksOf().flatMap((block) => {
        const preview = block.querySelector("[data-preview]");
        if (!preview) return [];
        const o = new MutationObserver((records) => {
            const swapped = records.some((r) =>
                [...r.addedNodes].some((n) => n.nodeType === Node.ELEMENT_NODE),
            );
            if (swapped) applyAllOptions();
        });
        o.__target = preview;
        return [o];
    });
    observePreview();
});
// applyAllOptions here as well as from the MutationObserver. The observer fires
// on the swap, but htmx settles afterwards and can replace the subtree again -
// so the options were being written onto markup that was then thrown away. That
// is why Tabs lost its "Panel surface" setting on every tab change.
document.body?.addEventListener("htmx:afterSettle", () => {
    applyAllOptions();
    syncCode();
    syncDesignSystem();
    syncMotion();
    syncOptionValues();
});

// Shared by the header button and Ctrl+T. A design system with only one scheme
// has nothing to toggle to, so this is a no-op there rather than a swap to a
// theme whose tokens the loaded stylesheet never defines.
function toggleTheme() {
    if (schemesFor(document.documentElement.dataset.ds).length === 1) return;
    const next =
        document.documentElement.dataset.theme === "dark" ? "light" : "dark";
    const swap = () =>
        withoutTransitions(() => {
            document.documentElement.dataset.theme = next;
        });
    // The new theme is wiped down over the old one from the top of the viewport,
    // so the swap reads as a deliberate change rather than as a flash.
    //
    // Gated on data-motion, not on the media query: the attribute is already the
    // OS preference with the site's Animation switch layered over it. The
    // ::view-transition pseudo-elements also sit outside the blanket "zero every
    // duration" rule in tokens.css - they are not descendants of <html> - so with
    // motion off this is the only thing that can stop the wipe.
    if (
        document.startViewTransition &&
        document.documentElement.dataset.motion !== "off"
    ) {
        document.startViewTransition(swap);
    } else {
        swap();
    }
    try {
        localStorage.setItem("qompnt-theme", next);
    } catch (_) {}
}

document.addEventListener("click", async (e) => {
    if (e.target.closest("[data-theme-toggle]")) {
        toggleTheme();
        return;
    }

    // Search: the icon button becomes the field, the clear button undoes it.
    const openSearch = e.target.closest("[data-search-open]");
    if (openSearch) {
        // No room for a field on a phone, so the same button opens the palette. The
        // width is checked at click time rather than at load: a rotated phone or a
        // resized window would otherwise keep whichever answer was right at boot.
        //
        // Only the one in the page header. Expanding Input is a component with this
        // exact hook, and its demo has to keep expanding at every width - it is what
        // the page is there to show.
        if (
            openSearch.closest("[data-sticky-header]") &&
            matchMedia("(max-width: 639px)").matches
        ) {
            openPalette();
            return;
        }
        const box = openSearch.closest("[data-search]");
        openSearch.hidden = true;
        const field = box.querySelector("[data-search-field]");
        field.hidden = false;
        field.querySelector("input").focus();
        return;
    }
    const clearSearch = e.target.closest("[data-search-clear]");
    if (clearSearch) {
        const box = clearSearch.closest("[data-search]");
        const input = box.querySelector('input[type="search"]');
        input.value = "";
        // Restores the full stack. Nothing is fetched - filterCards only unhides
        // what is already there. Header only: Expanding Input ships this same
        // clear button, and its demo is rendered inside a card on the index, so
        // an unscoped call would reset the page's filter from inside a preview.
        if (clearSearch.closest("[data-sticky-header]")) filterCards("");
        box.querySelector("[data-search-field]").hidden = true;
        box.querySelector("[data-search-open]").hidden = false;
        return;
    }

    const copySrc = e.target.closest("[data-copy-src]");
    if (copySrc) {
        try {
            const res = await fetch(copySrc.dataset.copySrc);
            await navigator.clipboard.writeText(await res.text());
            copySrc.setAttribute("aria-label", "Copied");
        } catch (_) {
            copySrc.setAttribute("aria-label", "Copy failed");
        }
        setTimeout(
            () => copySrc.setAttribute("aria-label", "Copy markup"),
            1600,
        );
        return;
    }

    // Copies a named code pane without a round trip: the panes are already in the
    // DOM, so the CSS button reads the same text the CSS tab shows rather than
    // fetching a second copy that could drift from it.
    const pane = e.target.closest("[data-copy-pane]");
    if (pane) {
        const idx = { html: 0, css: 1, tw: 2 }[pane.dataset.copyPane] ?? 0;
        const code = document
            .querySelectorAll(".panes > .panels > pre")
            [idx]?.querySelector("code");
        try {
            await navigator.clipboard.writeText(code.textContent);
            pane.setAttribute("aria-label", "Copied");
        } catch (_) {
            pane.setAttribute("aria-label", "Copy failed");
        }
        setTimeout(() => pane.setAttribute("aria-label", "Copy CSS"), 1600);
        return;
    }

    const copy = e.target.closest("[data-copy-code]");
    if (copy) {
        const shown = [
            ...document.querySelectorAll(".panes > .panels > pre"),
        ].find((el) => getComputedStyle(el).display !== "none");
        const code = (
            shown || document.querySelector("[data-code]")
        )?.querySelector("code");
        try {
            await navigator.clipboard.writeText(code.textContent);
            copy.setAttribute("aria-label", "Copied");
        } catch (_) {
            copy.setAttribute("aria-label", "Copy failed");
        }
        setTimeout(() => copy.setAttribute("aria-label", "Copy code"), 1600);
        return;
    }

    // Combobox: a click on an option fills the input; a click on the input opens.
    const cbOpt = e.target.closest('[data-combobox] [role="option"]');
    if (cbOpt) {
        const box = cbOpt.closest("[data-combobox]");
        const value = cbOpt.textContent.trim();
        box.querySelector('input[role="combobox"]').value = value;
        box.querySelector('input[type="hidden"]').value = value;
        cbOptions(box).forEach((o) =>
            o.setAttribute("aria-selected", String(o === cbOpt)),
        );
        box.querySelector('[role="listbox"]').hidden = true;
        return;
    }
    const cbInput = e.target.closest('[data-combobox] input[role="combobox"]');
    if (cbInput)
        cbInput
            .closest("[data-combobox]")
            .querySelector('[role="listbox"]').hidden = false;
    document.querySelectorAll("[data-combobox]").forEach((box) => {
        if (!box.contains(e.target))
            box.querySelector('[role="listbox"]').hidden = true;
    });

    // Custom select: pick an option, mirror it into the trigger and the hidden
    // input, close. Clicking anywhere else closes any open one.
    const opt = e.target.closest('[data-select] [role="option"]');
    if (opt) {
        const d = opt.closest("details");
        const label = opt.textContent.trim();
        // data-value is the submitted value when it differs from the label - and an
        // empty data-value is meaningful ("no filter"), so test for the attribute.
        const value = opt.hasAttribute("data-value")
            ? opt.dataset.value
            : label;
        d.querySelector("[data-select-label]").textContent = label;
        const hidden = d.querySelector('input[type="hidden"]');
        if (hidden) hidden.value = value;
        d.querySelectorAll('[role="option"]').forEach((o) =>
            o.setAttribute("aria-selected", String(o === opt)),
        );
        d.open = false;
        if (opt.hasAttribute("data-ds-option"))
            applyDesignSystem(opt.dataset.dsOption);
        return;
    }
    document.querySelectorAll("details[data-select][open]").forEach((d) => {
        if (!d.contains(e.target)) d.open = false;
    });
});

// ---- Command palette -------------------------------------------------------
//
// The whole component list is already in the page (see "palette" in
// layout.html), so this filters rows rather than querying anything. Selection is
// tracked as an index into the visible rows instead of by focusing them: the
// input has to keep focus for typing to work, and arrow keys have to move the
// highlight without moving the caret out of the field.
//
// ponytail: substring match on the name. Fuzzy matching goes in when a name is
// long enough that people mistype it - thirty short names is not that.
const palette = {
    get el() {
        return document.querySelector("[data-palette]");
    },
    index: 0,
};

function paletteRows() {
    return [...palette.el.querySelectorAll("[data-palette-item]")].filter(
        (r) => !r.hidden,
    );
}

function paletteSelect(i) {
    const rows = paletteRows();
    if (!rows.length) return;
    // Wraps, so holding Down does not stick at the bottom of a list this short.
    palette.index = (i + rows.length) % rows.length;
    rows.forEach((r, n) =>
        r.setAttribute("aria-selected", String(n === palette.index)),
    );
    rows[palette.index].scrollIntoView({ block: "nearest" });
}

function paletteFilter(q) {
    const needle = q.trim().toLowerCase();
    const el = palette.el;
    let shown = 0;
    el.querySelectorAll("[data-palette-item]").forEach((r) => {
        r.hidden =
            needle !== "" && !r.dataset.name.toLowerCase().includes(needle);
        if (!r.hidden) shown++;
    });
    el.querySelector("[data-palette-empty]").hidden = shown > 0;
    el.querySelector("[data-palette-heading]").hidden = shown === 0;
    paletteSelect(0);
}

function openPalette() {
    const el = palette.el;
    if (!el || el.open) return;
    el.showModal();
    const input = el.querySelector("[data-palette-input]");
    input.value = "";
    paletteFilter("");
    input.focus();
}

document.addEventListener("input", (e) => {
    if (e.target.matches("[data-palette-input]")) paletteFilter(e.target.value);
});

document.addEventListener("keydown", (e) => {
    // Ctrl/Cmd+K opens, Ctrl/Cmd+T toggles the theme. Both are claimed from the
    // browser: Cmd+K is a Chrome address-bar shortcut and Ctrl+T opens a tab, so
    // preventDefault is the point, not a formality.
    const mod = e.metaKey || e.ctrlKey || e.altKey;
    if (mod && e.key.toLowerCase() === "k") {
        e.preventDefault();
        openPalette();
        return;
    }
    if (mod && e.key.toLowerCase() === "t") {
        e.preventDefault();
        toggleTheme();
        return;
    }

    const el = palette.el;
    if (!el?.open) return;
    if (e.key === "ArrowDown") {
        e.preventDefault();
        paletteSelect(palette.index + 1);
    } else if (e.key === "ArrowUp") {
        e.preventDefault();
        paletteSelect(palette.index - 1);
    } else if (e.key === "Enter") {
        e.preventDefault();
        paletteRows()[palette.index]?.click();
    }
    // Escape is <dialog>'s own, and it already does the right thing.
});

// Clicking the backdrop closes. The dialog fills its own box, so a click whose
// target is the dialog element itself landed outside the content.
document.addEventListener("click", (e) => {
    if (e.target.matches("[data-palette]")) e.target.close();
    // htmx boosts navigation, so following a row swaps the page under an open
    // dialog and leaves it there.
    if (e.target.closest("[data-palette-item]")) palette.el.close();
});

document.addEventListener("click", (e) => {
    if (e.target.closest("[data-palette-open]")) openPalette();
});

// The sticky header separates itself by blurring what passes under it, which is
// pure CSS - see [data-sticky-header] in tokens.css. There used to be an
// IntersectionObserver here toggling a border on once the header was stuck; the
// blur does that job at every scroll position, so it is gone.

// ---- Index search ----------------------------------------------------------
//
// Filters the cards already in the page. This used to be an htmx GET to
// /p/search on every keystroke, which meant a server round trip - and on Vercel
// a function invocation billed for CPU - per character typed. The whole list is
// thirty cards and it is already rendered, so the network was never buying
// anything the browser could not do itself.
//
// The count badge is filtered too, because it always claimed to count what was
// on screen.
function filterCards(q) {
    const needle = q.trim().toLowerCase();
    const cards = document.querySelectorAll("[data-card]");
    let shown = 0;
    cards.forEach((c) => {
        const hit = needle === "" || c.dataset.name.toLowerCase().includes(needle);
        c.hidden = !hit;
        if (hit) shown++;
    });
    const empty = document.querySelector("[data-cards-empty]");
    if (empty) empty.hidden = shown > 0 || cards.length === 0;
    const count = document.querySelector("[data-card-count]");
    if (count) count.textContent = String(shown);
}

document.addEventListener("input", (e) => {
    if (e.target.matches("[data-search-input]")) filterCards(e.target.value);
});

// ---- Level Slider: how many stops --------------------------------------------
//
// The generic option handlers swap classes or attributes; this one has to move
// several things at once - the range's max, its value if that max just passed
// it, the fill, and where each remaining label sits - so it gets a handler of
// its own rather than a data-opt-* spelling that would only ever fit this
// control.
//
// The markup always ships six stops. Fewer is the six with the tail hidden and
// the rest respaced, which keeps this reversible: nothing is removed, so sliding
// back up restores the labels the component came with.
function setLevelStops(level, n) {
    const stops = [...level.querySelectorAll(".level-stops > *")];
    stops.forEach((s, i) => {
        s.hidden = i >= n;
        // Fractions of the track, so the last stop lands exactly on the end.
        if (i < n) s.style.setProperty("--stop", n === 1 ? 0 : i / (n - 1));
    });

    const range = level.querySelector('input[type="range"]');
    if (!range) return;
    range.max = String(n - 1);
    if (Number(range.value) > n - 1) range.value = String(n - 1);
    range.style.setProperty("--p", (Number(range.value) / (n - 1)) * 100);
    level.dataset.v = range.value;
}

document.addEventListener("input", (e) => {
    const ctl = e.target.closest("[data-opt-levels]");
    if (!ctl) return;
    const level = ctl
        .closest("[data-block]")
        ?.querySelector("[data-preview] .level");
    if (level) setLevelStops(level, Number(ctl.value));
});

// ---- Keyboard hints ---------------------------------------------------------
//
// The shortcuts are Cmd on a Mac and Ctrl everywhere else - the handlers already
// accept either - so the labels have to say which. Written as Ctrl in the markup
// and rewritten here, rather than the reverse, because Ctrl is what the larger
// share of visitors will press and it is the correct answer if this never runs.
//
// navigator.platform is deprecated but is still the only reliable Mac signal;
// userAgentData.platform is not exposed everywhere, hence both.
const isApplePlatform = /mac|iphone|ipad|ipod/i.test(
    navigator.userAgentData?.platform || navigator.platform || navigator.userAgent,
);

function localiseShortcuts() {
    if (!isApplePlatform) return;
    document.querySelectorAll("[data-kbd-mod]").forEach((el) => {
        el.textContent = "⌘";
    });
    document.querySelectorAll("[data-tip]").forEach((el) => {
        if (el.dataset.tip.includes("Ctrl")) {
            el.dataset.tip = el.dataset.tip.replace("Ctrl", "⌘");
        }
    });
}
localiseShortcuts();
// hx-boost replaces the header, and with it the tooltips just rewritten.
document.body?.addEventListener("htmx:afterSettle", localiseShortcuts);

// ---- Restoring preferences on a back/forward navigation ----------------------
//
// The theme, the design system and the motion setting are written onto <html> by
// an inline script in the head, which runs once per real page load. A back
// gesture usually does not do a real page load: the browser restores the page
// whole out of its back/forward cache, DOM and all, and that DOM is the one it
// froze - with whatever theme was on it at the time.
//
// So changing the theme on a component page and swiping back showed the page
// still in the old theme, while the back *button* often looked correct because
// it more often does a fresh load. Same for the design system and motion.
//
// pageshow with persisted set is the event for exactly this: it fires only on a
// cached restore, and it is the one place the head script cannot reach.
function restorePreferences() {
    const root = document.documentElement;
    let theme, ds, motion;
    try {
        theme = localStorage.getItem("qompnt-theme");
        ds = localStorage.getItem("qompnt-ds");
        motion = localStorage.getItem("qompnt-motion");
    } catch (_) {
        return;
    }

    if (theme) root.dataset.theme = theme;
    if (motion) root.dataset.motion = motion;

    // Without transitions: this is a restore, not a change the visitor made, and
    // a page fading between themes as it comes back reads as a bug.
    withoutTransitions(() => {
        if ((ds || "") !== (root.dataset.ds || "")) applyDesignSystemLink(ds || "");
    });

    syncDesignSystem();
    syncMotion();
    localiseShortcuts();
}

// The stylesheet half of applyDesignSystem, without the persisting - the value
// being restored came out of storage in the first place.
function applyDesignSystemLink(name) {
    const link = document.getElementById("ds");
    if (link && name) link.href = `/static/themes/${name}.css?v=${assetV}`;
    else if (link) link.removeAttribute("href");
    document.documentElement.dataset.ds = name || "";
}

window.addEventListener("pageshow", (e) => {
    if (e.persisted) restorePreferences();
});

// The other way a page comes back without reloading: hx-boost handles the click,
// so Back is a history entry htmx restores from its own snapshot rather than a
// navigation the browser performs. Same failure, same fix.
document.body?.addEventListener("htmx:historyRestore", restorePreferences);
