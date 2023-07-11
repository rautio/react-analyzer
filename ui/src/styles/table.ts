import { css } from "lit";
export const table = css`
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
export default table;
