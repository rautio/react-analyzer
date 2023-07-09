import { LitElement, html, css } from "lit";
import { customElement } from "lit/decorators.js";
import report from "../report.json";
import "./export";

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
  render() {
    console.log(report);
    console.log(report.import_graph.edges.sort((a,b) => {
      if (a.id > b.id) return 1;
      return -1;
    }))
    return html`
      <div class="wrapper">
        <h2>Files</h2>
        <div class="summary">Total: ${report.summary.file_count} files</div>
        <div>
          ${report.exports.map(
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
