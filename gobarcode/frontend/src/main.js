import './style.css';
import './app.css';

import logo from './assets/images/logo-universal.png';
import { Greet, SelectFile } from '../wailsjs/go/main/App';

const filebtn = document.getElementById("filebtn")

filebtn.addEventListener("click", SelectFile)
