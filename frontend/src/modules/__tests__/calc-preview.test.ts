import { describe, it, expect, beforeEach, vi } from 'vitest';
import { CalcPreview } from '../calc-preview';

// Mock the Wails backend — evaluate simple arithmetic locally so tests run
// without a real Wails runtime.
vi.mock('../../../wailsjs/go/main/App', () => ({
    EvalCalc: vi.fn(async (expr: string): Promise<string> => {
        // Mirror the backend: require at least one operator (IsCalcQuery behaviour)
        if (!/[+\-*/^%()]/.test(expr) || !/\d/.test(expr)) return '';
        try {
            // eslint-disable-next-line no-new-func
            const result = new Function('"use strict"; return (' + expr + ')')() as number;
            if (typeof result !== 'number' || !isFinite(result)) return '';
            // Mirror the backend: trim trailing zeros on decimals
            return parseFloat(result.toPrecision(10)).toString();
        } catch {
            return '';
        }
    }),
}));

describe('CalcPreview', () => {
    let el: HTMLElement;
    let preview: CalcPreview;

    beforeEach(() => {
        el = document.createElement('div');
        preview = new CalcPreview(el);
    });

    it('shows result for simple addition', async () => {
        await preview.update('1 + 2');
        expect(el.textContent).toBe('= 3');
    });

    it('shows result for subtraction', async () => {
        await preview.update('10 - 4');
        expect(el.textContent).toBe('= 6');
    });

    it('shows result for multiplication', async () => {
        await preview.update('4 * 5');
        expect(el.textContent).toBe('= 20');
    });

    it('shows result for division', async () => {
        await preview.update('15 / 3');
        expect(el.textContent).toBe('= 5');
    });

    it('shows a trimmed decimal result', async () => {
        await preview.update('10 / 3');
        // Should be a decimal number, not show trailing zeros
        expect(el.textContent).toMatch(/^= \d+\.\d+$/);
        expect(el.textContent).not.toMatch(/0+$/);
    });

    it('respects operator precedence', async () => {
        await preview.update('2 + 3 * 4');
        expect(el.textContent).toBe('= 14');
    });

    it('clears for non-numeric input', async () => {
        await preview.update('hello world');
        expect(el.textContent).toBe('');
    });

    it('clears for a number without operators', async () => {
        await preview.update('42');
        expect(el.textContent).toBe('');
    });

    it('clears after having shown a result', async () => {
        await preview.update('1 + 1');
        await preview.update('');
        expect(el.textContent).toBe('');
    });

    it('sets aria-hidden to false when showing a result', async () => {
        await preview.update('2 + 2');
        expect(el.getAttribute('aria-hidden')).toBe('false');
    });

    it('sets aria-hidden to true when cleared via clear()', () => {
        preview.clear();
        expect(el.getAttribute('aria-hidden')).toBe('true');
    });

    it('sets aria-hidden to true when input is non-math', async () => {
        await preview.update('not math');
        expect(el.getAttribute('aria-hidden')).toBe('true');
    });
});
