import * as _ from "underscore";
import { foo as bar } from "foobar";
import a from "b";

function foo() {
  let aList = [1, 2, 3, 4];
  const mappedList = _.map(aList, i => bar.add(i) + 1);
  a(mappedList);
}
