import { JSONEditor } from "vanilla-jsoneditor";
import { useEffect, useRef } from "react";

export function VanillaJSONEditor(props) {
  const refContainer = useRef(null);
  const refEditor = useRef(null);

  useEffect(() => {
    // create editor
    console.log("create editor", refContainer.current);
    refEditor.current = new JSONEditor({
      target: refContainer.current,
      props: {}
    });

    return () => {
      // destroy editor
      if (refEditor.current) {
        console.log("destroy editor");
        refEditor.current.destroy();
        refEditor.current = null;
      }
    };
  }, []);

  // update props
  useEffect(() => {
    if (refEditor.current) {
      console.log("update props", props);
      refEditor.current.updateProps(props);
    }
  }, [props]);

  return <div style={{height:"55vh"}} className="vanilla-jsoneditor-react" ref={refContainer}></div>;
}
