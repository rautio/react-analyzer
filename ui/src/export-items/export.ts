import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";
import { Export, Targets } from "./index";
import { table } from "../styles/table";

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
    ${table}
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
