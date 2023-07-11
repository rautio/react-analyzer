import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";
import { table } from "./styles/table";

export type PackageJson = {
  dependencies: Record<string, number>;
};

@customElement("package-json")
export class PJson extends LitElement {
  static styles = css`
    ${table}
  `;

  @property({ type: Object })
  package_json: PackageJson = { dependencies: {} };

  constructor() {
    super();
  }

  render() {
    return html`<div>
      <h2>Dead Files</h2>
      <table>
        <tr>
          <th>NPM</th>
          <th># of times used</th>
        </tr>
        ${Object.keys(this.package_json.dependencies).map(
          (dep) =>
            html`<tr>
              <td>${dep}</td>
              <td class="count">${this.package_json.dependencies[dep]}</td>
            </tr>`
        )}
      </table>
    </div>`;
  }
}
