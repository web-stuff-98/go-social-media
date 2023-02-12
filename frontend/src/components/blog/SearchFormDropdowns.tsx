import Dropdown from "../shared/Dropdown";
import classes from "../../styles/pages/Blog.module.scss";
import usePosts from "../../context/PostsContext";

export default function SearchFormDropdowns() {
  const {
    getSortModeFromParams,
    getSortOrderFromParams,
    setSortModeInParams,
    setSortOrderInParams,
  } = usePosts();

  return (
    <div
      data-testid="Search form dropdowns"
      className={classes.dropdownsContainer}
    >
      <Dropdown
        light
        index={getSortOrderFromParams === "DESCENDING" ? 0 : 1}
        setIndex={setSortOrderInParams}
        items={[
          { name: "DESCENDING", node: "Desc" },
          { name: "ASCENDING", node: "Asc" },
        ]}
      />
      <div className={classes.sortMode}>
        <Dropdown
          light
          index={getSortModeFromParams === "DATE" ? 0 : 1}
          setIndex={setSortModeInParams}
          items={[
            { name: "DATE", node: "Date" },
            { name: "POPULARITY", node: "Popularity" },
          ]}
        />
      </div>
    </div>
  );
}
