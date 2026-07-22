import './style.css';
import './app.css';


import * as go from '../wailsjs/go/main/App';

const filebtn = document.getElementById("filebtn");
const workName = document.getElementById("sheetname");
const setRow = document.getElementById("btnsetRow");
filebtn.addEventListener("click", SetFile);
setRow.addEventListener("click", SetHeader);

async function SetFile() {
  try {
    const fileInfo = await go.SelectFile()
    workName.innerText = fileInfo.selected_sheet_name
  } catch (error) {
    console.log(error)
  }

}

async function SetHeader() {
  const text = document.getElementById("headerinput").value;
  try {
    const headerinfo = await go.GetHeaders(Number(text))
    for (let index = 0; index < headerinfo.length; index++) {
      console.log("reading row: ", index + 1)
      console.log(headerinfo[index])

    }
  } catch (error) {
    console.log(error)
  }
}
