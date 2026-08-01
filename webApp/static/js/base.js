let apiPort = "";

function showFieldErrors(errors) {
    clearFieldErrors();
    if (!errors) return;
    for (let field in errors) {
        let el = document.getElementById("err-" + field);
        if (el) {
            el.innerText = errors[field];
            el.classList.remove("d-none");
        }
    }
}

function clearFieldErrors() {
    document.querySelectorAll(".field-error").forEach(el => {
        el.innerText = "";
        el.classList.add("d-none");
    });
}

function logout() {
    fetch(apiPort + "/api/logout", {
        method: "POST",
    })
    .then(() => {
        location.href = "/login";
    });
}

// Search
(function() {
    const searchInput = document.getElementById("book-search");
    const searchResults = document.getElementById("search-results");
    let debounceTimer = null;
    let abortController = null;
    let lastQuery = "";
    let activeIndex = -1;

    searchInput.addEventListener("focus", function() {
        if (searchResults.innerHTML && searchInput.value.trim().length >= 3) {
            searchResults.classList.add("show");
        }
    });

    searchInput.addEventListener("input", function() {
        clearTimeout(debounceTimer);
        const q = searchInput.value.trim();

        if (q.length < 3) {
            if (abortController) abortController.abort();
            searchResults.classList.remove("show");
            searchResults.innerHTML = "";
            lastQuery = "";
            activeIndex = -1;
            return;
        }

        if (q === lastQuery) {
            // Avoid fetching again if the trimmed query hasn't changed
            if (searchResults.innerHTML) {
                 searchResults.classList.add("show");
            }
            return;
        }

        debounceTimer = setTimeout(() => {
            if (abortController) abortController.abort();
            abortController = new AbortController();
            lastQuery = q;

            fetch(apiPort + "/api/search?q=" + encodeURIComponent(q), {
                        signal: abortController.signal,
            })
            .then(r => {
                if (!r.ok) {
                    throw new Error("Search failed");
                }
                return r.json();
            })
            .then(items => {
                searchResults.innerHTML = "";
                searchResults.classList.add("show");
                activeIndex = -1;

                if (!items || items.length === 0) {
                    searchResults.innerHTML = '<span class="dropdown-item text-muted">No result for search</span>';
                    return;
                }

                let html = '';
                items.forEach(item => {
                    html += '<a href="' + escapeHtml(item.url) + '" class="dropdown-item d-flex justify-content-between align-items-center">'
                        + escapeHtml(item.name)
                        + '<span class="badge bg-secondary">' + escapeHtml(item.kind) + '</span>'
                        + '</a>';
                });
                searchResults.innerHTML = html;
            })
            .catch(err => {
                if (err.name !== "AbortError") {
                    searchResults.classList.remove("show");
                }
            });
        }, 200);
    });

    searchInput.addEventListener("keydown", function(e) {
        if (e.key === "Escape") {
            searchResults.classList.remove("show");
            searchInput.blur();
            return;
        }

        const items = searchResults.querySelectorAll("a.dropdown-item");
        if (!searchResults.classList.contains("show") || items.length === 0) return;

        if (e.key === "ArrowDown") {
            e.preventDefault();
            activeIndex = (activeIndex + 1) % items.length;
            updateActive(items);
        } else if (e.key === "ArrowUp") {
            e.preventDefault();
            if (activeIndex <= 0) {
                activeIndex = items.length - 1;
            } else {
                activeIndex--;
            }
            updateActive(items);
        } else if (e.key === "Enter") {
            if (activeIndex >= 0) {
                e.preventDefault();
                items[activeIndex].click();
            }
        }
    });

    function updateActive(items) {
        items.forEach((item, index) => {
            if (index === activeIndex) {
                item.classList.add("focused");
            } else {
                item.classList.remove("focused");
            }
        });
    }

    document.addEventListener("click", function(e) {
        if (!document.getElementById("search-wrapper").contains(e.target)) {
            searchResults.classList.remove("show");
        }
    });

    function escapeHtml(text) {
        if (!text) return "";
        return String(text).replace(/[&<>"']/g, function(m) {
            return {
                '&': '&amp;',
                '<': '&lt;',
                '>': '&gt;',
                '"': '&quot;',
                "'": '&#39;'
            }[m];
        });
    }
})();
