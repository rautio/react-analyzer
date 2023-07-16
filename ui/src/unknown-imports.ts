import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";

@customElement("unknown-imports")
export class UnknownImports extends LitElement {
  static styles = css``;

  @property({ type: Array })
  unknown_imports: Array<String> = [];

  constructor() {
    super();
  }

  render() {
    return html`<div>
      <h2>Unknown Imports</h2>
      <ul>
        ${this.unknown_imports.sort((a,b) => {
          if (a > b) return 1;
          return -1;
        }).map((file) => html`<li>${file}</li>`)}
      </ul>
    </div>`;
  }
}
