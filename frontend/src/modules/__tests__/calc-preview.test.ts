import { describe, it, expect, beforeEach } from 'vitest';
import { CalcPreview } from '../calc-preview';

describe('CalcPreview', () => {
    let el: HTMLElement;
    let preview: CalcPreview;

    beforeEach(() => {
        el = document.createElement('div');
        preview = new CalcPreview(el);
    });

    it('shows the provided backend result', () => {
        preview.show('3');
        expect(el.textContent).toBe('= 3');
    });

    it('clears after having shown a result', () => {
        preview.show('2');
        preview.clear();
        expect(el.textContent).toBe('');
    });

    it('sets aria-hidden to false when showing a result', () => {
        preview.show('4');
        expect(el.getAttribute('aria-hidden')).toBe('false');
    });

    it('sets aria-hidden to true when cleared via clear()', () => {
        preview.clear();
        expect(el.getAttribute('aria-hidden')).toBe('true');
    });

    it('starts without a value until shown or cleared', () => {
        expect(el.textContent).toBe('');
    });
});
