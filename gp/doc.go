/*
Package gp is a library for generic programming in Go providing core types and methods.

Values

The Value type is defined as an empty interface hence any type can be used for a specific model.
Note that typed GP is not yet supported, it's assumed all values will be of the same type 
for a given data set. See gogp/num for an example of implementing a floating point numeric type.

Opcodes

The basic atom of the model is the Opcode interface. The implemention must supply the Eval 
method to evalute an opcode, an Arity method to define the number of arguments and a Format 
and a String method for printing. 

Expressions

The Expr type is defined a slice of Opcodes. An expression is stored internally as a list in prefix
notation to represent the opcode tree. Methods are provided to evaluate an expression given specified
terminal nodes.

Individuals

An Individual has a code expression which represents the genome and a fitness value as calculated by
the implementation of the Evaluator interface. The fitness should be a normalised floating point
value in the range 0 to 1 where 0 is the worst possible and 1 represents a perfect solution to the 
problem.

Populations

A Population is a slice of individuals. Implementations of the Selection interface are 
provided to pick a subset from the population, and of the Variation interface to provide mutation
and crossover genetic operators. Decorators can be used to further customise a Variation, e.g. for
size limits for bloat control.
*/
package gp


