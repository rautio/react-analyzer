import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";

@customElement("dead-files")
export class DeadFiles extends LitElement {
  static styles = css``;

  @property({ type: Array })
  dead_files: Array<String> = [];

  constructor() {
    super();
  }

  render() {
    return html`<div>
      <h2>Dead Files</h2>
      <ul>
        ${this.dead_files.map((file) => html`<li>${file}</li>`)}
      </ul>
    </div>`;
  }
}
