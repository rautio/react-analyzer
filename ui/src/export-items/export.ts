import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";

interface Export {
  is_default: boolean;
  name: string;
  target: string;
}

interface Targets {
  count: number;
  is_default: boolean;
  targets: Array<string>;
}

const parseExports = (exports: Array<Export>) => {
  // Name -> # of targets, raw targets
  const exportMap: Record<string, Targets> = {};
  exports.forEach((ex) => {
    if (ex.name in exportMap) {
      exportMap[ex.name].count++;
      exportMap[ex.name].is_default = ex.is_default;
      exportMap[ex.name].targets.push(ex.target);
    } else {
      exportMap[ex.name] = {
        count: 1,
        is_default: ex.is_default,
        targets: [ex.target],
      };
    }
  });
  return exportMap;
};

@customElement("export-item")
export class ExportItem extends LitElement {
  static styles = css`
    table {
      margin: 10px;
      margin-left: 20px;
    }
    table,
    td,
    th {
      border-collapse: collapse;
      border: 1px solid;
      border-color: #063665;
    }
    th {
      background-color: #063665;
    }
    td {
      background-color: #074989;
    }
    td.count {
      text-align: center;
    }
    td {
      padding-left: 1em;
      padding-right: 1em;
    }
  `;
  @property({ type: Array<Export> })
  exports = [];

  constructor() {
    super();
  }

  render() {
    const data = parseExports(this.exports);
    const exportNames = Object.keys(data);
    return html`<table>
      <tr>
        <th>Name</th>
        <th># of times imported</th>
      </tr>
      ${exportNames.map(
        (ex) =>
          html`<tr>
            <td>${ex}</td>
            <td class="count">
              <span title=${data[ex].targets.join("\n")}
                >${data[ex].count}</span
              >
            </td>
          </tr>`
      )}
      <table></table>
    </table>`;
  }
}
