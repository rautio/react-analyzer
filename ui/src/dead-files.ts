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
      <h2>Unimported Files</h2>
      <ul>
        ${this.dead_files.sort((a,b) => {
          if (a > b) return 1;
          return -1;
        }).map((file) => html`<li>${file}</li>`)}
      </ul>
    </div>`;
  }
}
