describe('Search flow', () => {
  it('shows suggestions and keyboard navigation works', () => {
    cy.visit('/');
    cy.get('input[aria-label="Search"]').should('exist').as('searchInput');
    cy.get('@searchInput').type('sare');
    cy.wait(500); // wait for debounce + network
    cy.get('ul').should('be.visible');
    cy.get('ul li').its('length').should('be.gte', 1);
    // Press arrow down and enter
    cy.get('@searchInput').type('{downarrow}{enter}');
    // We expect navigation to have occurred (URL changed or new page)
    cy.location('pathname').should('not.equal', '/');
  });
});
