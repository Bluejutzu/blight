import { escapeHtml } from './utils';

const STORAGE_KEY = 'blight-search-history';
const MAX_HISTORY = 10;

export class SearchHistory {
    private containerEl: HTMLElement;
    private searchInputEl: HTMLInputElement;
    private onSelect: (query: string) => void;
    private history: string[];
    private highlightedIndex = -1;

    constructor(
        containerEl: HTMLElement,
        searchInputEl: HTMLInputElement,
        onSelect: (query: string) => void
    ) {
        this.containerEl = containerEl;
        this.searchInputEl = searchInputEl;
        this.onSelect = onSelect;
        this.history = JSON.parse(localStorage.getItem(STORAGE_KEY) || '[]');
    }

    add(query: string): void {
        if (!query || query.length < 2) return;
        this.history = [query, ...this.history.filter((q) => q !== query)].slice(0, MAX_HISTORY);
        localStorage.setItem(STORAGE_KEY, JSON.stringify(this.history));
    }

    show(): void {
        if (this.history.length === 0) {
            this.containerEl.classList.add('hidden');
            return;
        }
        this.highlightedIndex = -1;
        this._render();
        this.containerEl.classList.remove('hidden');

        this.searchInputEl.setAttribute('aria-expanded', 'true');
    }

    hide(): void {
        this.highlightedIndex = -1;
        this.containerEl.classList.add('hidden');
        this.searchInputEl.setAttribute('aria-expanded', 'false');
    }

    isVisible(): boolean {
        return !this.containerEl.classList.contains('hidden');
    }

    /** Move highlight down; returns true if the history dropdown consumed the event. */
    navigateDown(): boolean {
        if (!this.isVisible()) return false;
        this.highlightedIndex = Math.min(this.highlightedIndex + 1, this.history.length - 1);
        this._updateHighlight();
        return true;
    }

    /** Move highlight up; returns true if the history dropdown consumed the event. */
    navigateUp(): boolean {
        if (!this.isVisible()) return false;
        if (this.highlightedIndex <= 0) {
            this.highlightedIndex = -1;
            this._updateHighlight();
            return false; // let the caller handle further up-arrows
        }
        this.highlightedIndex--;
        this._updateHighlight();
        return true;
    }

    /** If an item is highlighted, select it and return true; otherwise return false. */
    confirmHighlighted(): boolean {
        if (!this.isVisible() || this.highlightedIndex < 0) return false;
        const query = this.history[this.highlightedIndex];
        if (query === undefined) return false;
        this.searchInputEl.value = query;
        this.hide();
        this.onSelect(query);
        return true;
    }

    private _render(): void {
        this.containerEl.innerHTML =
            `<div class="history-header">Recent</div>` +
            this.history
                .map(
                    (q, i) => `
                <div class="history-item" data-index="${i}" role="option">
                    <span class="history-item-icon">↺</span>
                    <span class="history-item-text">${escapeHtml(q)}</span>
                    <span class="history-item-remove" data-remove="${i}" title="Remove">✕</span>
                </div>
            `
                )
                .join('');

        this.containerEl.querySelectorAll<HTMLElement>('.history-item').forEach((item) => {
            item.addEventListener('mousedown', (e) => {
                const remove = (e.target as HTMLElement).closest(
                    '[data-remove]'
                ) as HTMLElement | null;
                if (remove) {
                    e.preventDefault();
                    const idx = parseInt(remove.dataset['remove'] ?? '0', 10);
                    this.history.splice(idx, 1);
                    localStorage.setItem(STORAGE_KEY, JSON.stringify(this.history));
                    this.show();
                    return;
                }
                e.preventDefault();
                const idx = parseInt(item.dataset['index'] ?? '0', 10);
                this.searchInputEl.value = this.history[idx] ?? '';
                this.hide();
                this.onSelect(this.searchInputEl.value);
            });
        });
    }

    private _updateHighlight(): void {
        this.containerEl.querySelectorAll<HTMLElement>('.history-item').forEach((item, i) => {
            item.classList.toggle('history-item--active', i === this.highlightedIndex);
        });
        if (this.highlightedIndex >= 0) {
            const items = this.containerEl.querySelectorAll<HTMLElement>('.history-item');
            items[this.highlightedIndex]?.scrollIntoView({ block: 'nearest' });
        }
    }
}
