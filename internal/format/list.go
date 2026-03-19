package format

import (
	"fmt"
	"io"

	"github.com/saberapp/cli/internal/client"
)

// PrintCompanyList renders a single company list.
func PrintCompanyList(w io.Writer, list *client.CompanyList) {
	KV(w, [][2]string{
		{"ID:", list.ID},
		{"Name:", list.Name},
		{"Industries:", JoinStrings(list.Filter.Industries, "—")},
		{"Sizes:", JoinStrings(list.Filter.Sizes, "—")},
		{"Created:", list.CreatedAt},
	})
}

// PrintCompanyLists renders a table of company lists.
func PrintCompanyLists(w io.Writer, lists []client.CompanyList, total int) {
	tw := NewTabWriter(w)
	fmt.Fprintln(tw, "ID\tNAME\tCREATED")
	for _, l := range lists {
		fmt.Fprintf(tw, "%s\t%s\t%s\n", l.ID, TruncateString(l.Name, 40), l.CreatedAt)
	}
	FlushTable(tw)
	fmt.Fprintf(w, "\n%d of %d lists\n", len(lists), total)
}

// PrintCompanies renders a table of companies from a list.
func PrintCompanies(w io.Writer, companies []client.CompanyListCompany, total int) {
	tw := NewTabWriter(w)
	fmt.Fprintln(tw, "NAME\tDOMAIN\tINDUSTRY\tSIZE\tCOUNTRY")
	for _, c := range companies {
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\t%s\n",
			TruncateString(c.Name, 30),
			c.Domain,
			TruncateString(c.Industry, 25),
			c.Size,
			c.CountryCode,
		)
	}
	FlushTable(tw)
	fmt.Fprintf(w, "\n%d of %d companies\n", len(companies), total)
}

// PrintCompanySearchResults renders search result companies.
func PrintCompanySearchResults(w io.Writer, companies []client.CompanyListCompany, total int) {
	PrintCompanies(w, companies, total)
}

// PrintContactList renders a single contact list.
func PrintContactList(w io.Writer, list *client.ContactList) {
	keywords := list.Filters.Keywords
	if keywords == "" {
		keywords = "—"
	}
	KV(w, [][2]string{
		{"ID:", list.ID},
		{"Name:", list.Name},
		{"Company LinkedIn:", JoinStrings(list.Filters.CompanyLinkedInURLs, "—")},
		{"Titles:", JoinStrings(list.Filters.JobTitles, "—")},
		{"Keywords:", keywords},
		{"Countries:", JoinStrings(list.Filters.Countries, "—")},
		{"Total contacts:", fmt.Sprintf("%d", list.ContactCount)},
		{"Created:", list.CreatedAt},
	})
}

// PrintContactLists renders a table of contact lists.
func PrintContactLists(w io.Writer, lists []client.ContactList, total int) {
	tw := NewTabWriter(w)
	fmt.Fprintln(tw, "ID\tNAME\tCONTACTS\tCREATED")
	for _, l := range lists {
		fmt.Fprintf(tw, "%s\t%s\t%d\t%s\n", l.ID, TruncateString(l.Name, 40), l.ContactCount, l.CreatedAt)
	}
	FlushTable(tw)
	fmt.Fprintf(w, "\n%d of %d lists\n", len(lists), total)
}

// PrintContactSearchResults renders a table of contact search results.
func PrintContactSearchResults(w io.Writer, contacts []client.ContactSearchResult, count int) {
	tw := NewTabWriter(w)
	fmt.Fprintln(tw, "NAME\tROLE\tCOMPANY\tLOCATION")
	for _, c := range contacts {
		name := c.FullName
		if name == "" {
			name = c.FirstName + " " + c.LastName
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			TruncateString(name, 25),
			TruncateString(c.Role, 30),
			TruncateString(c.CompanyName, 25),
			c.Location,
		)
	}
	FlushTable(tw)
	fmt.Fprintf(w, "\n%d contacts found\n", count)
}

// PrintContacts renders a table of contacts in a list.
func PrintContacts(w io.Writer, contacts []client.ContactListItem, total int) {
	tw := NewTabWriter(w)
	fmt.Fprintln(tw, "NAME\tROLE\tCOMPANY\tLOCATION")
	for _, c := range contacts {
		name := c.FullName
		if name == "" {
			name = c.FirstName + " " + c.LastName
		}
		fmt.Fprintf(tw, "%s\t%s\t%s\t%s\n",
			TruncateString(name, 25),
			TruncateString(c.Role, 30),
			TruncateString(c.CompanyName, 25),
			c.Location,
		)
	}
	FlushTable(tw)
	fmt.Fprintf(w, "\n%d of %d contacts\n", len(contacts), total)
}
