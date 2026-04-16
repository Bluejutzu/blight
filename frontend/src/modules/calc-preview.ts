export class CalcPreview {
    private el: HTMLElement;

    constructor(el: HTMLElement) {
        this.el = el;
    }

    show(result: string): void {
        this.el.textContent = '= ' + result;
        this.el.setAttribute('aria-hidden', 'false');
    }

    clear(): void {
        this.el.textContent = '';
        this.el.setAttribute('aria-hidden', 'true');
    }
}
