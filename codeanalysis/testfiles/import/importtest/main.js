import foo from "foo";
import * as bar from "bar";
import { a } from "a";
import { e as b } from "b";
import { c1 , c2} from "c";
import { e1 , e2 as d1 , d2 } from "d";
import f1, * as f from "e";
import "f";
import 'g';

import "/f";
import "./f";
import "../f";

import foo1 from "/foo";
import foo2 from "./foo";
import foo3 from "../foo";

function main() {
    console.log("Hello World")
}