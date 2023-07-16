import { LitElement, html, css } from "lit";
import { customElement, property } from "lit/decorators.js";

export enum Nav {
  all,
  dead,
  unknown,
  pJson,
}

// type OnClick = (nav: Nav) => void;

@customElement("nav-header")
export class NavHeader extends LitElement {
  static styles = css`
    .nav-item {
      background: none;
      padding: 10px 20px;
      border: none;
      border-right: 1px solid;
      border-top: 1px solid;
      cursor: pointer;
    }
    .active {
      background-color: #063665;
    }
  `;

  @property({ type: Nav })
  active = Nav.all;

  // @property({ type: OnClick })
  @property()
  onClick = () => {};
  constructor() {
    super();
  }

  _handleClick = (nav: Nav) => {
    // @ts-ignore
    this.onClick(nav);
  };

  render() {
    return html`
      <nav>
        <button
          @click="${() => this._handleClick(Nav.all)}"
          class="nav-item ${this.active == Nav.all ? "active" : ""}"
        >
          All Files
        </button>
        <button
          @click="${() => this._handleClick(Nav.dead)}"
          class="nav-item ${this.active == Nav.dead ? "active" : ""}"
        >
          Dead Files
        </button>
        <button
          @click="${() => this._handleClick(Nav.unknown)}"
          class="nav-item ${this.active == Nav.unknown ? "active" : ""}"
        >
          Unknown Imports
        </button>
        <button
          @click="${() => this._handleClick(Nav.pJson)}"
          class="nav-item ${this.active == Nav.pJson ? "active" : ""}"
        >
          Package JSON
        </button>
        <hr style="margin-top: 0; " />
      </nav>
    `;
  }
}
