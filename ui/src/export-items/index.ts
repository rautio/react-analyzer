import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";
import "./export";

export interface Export {
  is_default: boolean;
  name: string;
  target: string;
}

export interface Targets {
  count: number;
  is_default: boolean;
  targets: Array<string>;
}

export type Summary = {
  file_count: number;
};

export interface Exports {
  source: String;
  exports: Array<Export>;
}

@customElement("export-items")
export class ExportItems extends LitElement {
  static styles = css`
    .wrapper {
      margin: 20px;
    }
    .summary {
      margin-bottom: 20px;
    }
  `;

  @property({ type: Array })
  exports: Array<Exports> = [];

  @property({ type: Object })
  summary: Summary = { file_count: 0 };

  render() {
    return html`
      <div class="wrapper">
        <h2>Files</h2>
        <div class="summary">Total: ${this.summary.file_count} files</div>
        <div>
          ${this.exports.map(
            (ex) =>
              html`<details>
                <summary>${ex.source}</summary>
                <export-item .exports=${ex.exports} />
              </details>`
          )}
        </div>
      </div>
    `;
  }
}
