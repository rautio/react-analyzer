import { LitElement, html } from 'lit';
import { customElement } from 'lit/decorators.js';
import report from '../report.json';

@customElement('export-items')
export class ExportItems extends LitElement {
  render() {
    return html`
      <div>
        <h2>Files</h2>
        <ul>
          ${report.exports.map((ex) => html`<li>${ex.source}</li>`)}
        </ul>
      </div>
    `;
  }
}